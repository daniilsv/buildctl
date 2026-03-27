package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/ai"
	"github.com/build-assistant/back/internal/git"
	"github.com/build-assistant/back/internal/notifications"
	"github.com/build-assistant/back/internal/ssh"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"log/slog"
)

type ArtifactService interface {
	GetBuildArtifactsAsInterface(ctx context.Context, buildID uuid.UUID) ([]interface{}, error)
}

type Processor struct {
	queries         *db.Queries
	gitClient       git.Client
	aiClient        ai.Client
	notifier        notifications.Notifier
	artifactService ArtifactService
	sshExecutor     *ssh.Executor
}

func NewProcessor(queries *db.Queries, gitClient git.Client, aiClient ai.Client, notifier notifications.Notifier, artifactService ArtifactService, sshExecutor *ssh.Executor) *Processor {
	return &Processor{
		queries:         queries,
		gitClient:       gitClient,
		aiClient:        aiClient,
		notifier:        notifier,
		artifactService: artifactService,
		sshExecutor:     sshExecutor,
	}
}

func (p *Processor) ProcessTask(ctx context.Context, task *Task) error {
	switch task.Type {
	case TaskTypeProcessSuccess:
		return p.processSuccess(ctx, task)
	case TaskTypeProcessFailed:
		return p.processFailed(ctx, task)
	case TaskTypeRedeploy:
		return p.processRedeploy(ctx, task)
	default:
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}

func (p *Processor) processSuccess(ctx context.Context, task *Task) error {
	projectID := task.Data["project_id"].(string)
	branchID := task.Data["branch_id"].(string)
	commitHash := task.Data["commit_hash"].(string)
	buildID := task.Data["build_id"].(string)

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	branchUUID, err := uuid.Parse(branchID)
	if err != nil {
		return fmt.Errorf("invalid branch ID: %w", err)
	}

	project, err := p.queries.GetProjectByID(ctx, projectUUID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	branch, err := p.queries.GetBranchByID(ctx, branchUUID)
	if err != nil {
		return fmt.Errorf("branch not found: %w", err)
	}

	lastCommitHash := ""
	if branch.LastSuccessfulCommit != nil {
		lastCommitHash = *branch.LastSuccessfulCommit
	}
	if lastCommitHash == "" {
		lastCommitHash = commitHash
	}

	commits, err := p.gitClient.GetCommits(ctx, &project, &branch, commitHash, lastCommitHash)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(commits) == 0 {
		return nil
	}

	authorName := commits[0].Author

	var summary string
	cached, err := p.queries.GetCommitsSummary(ctx, &db.GetCommitsSummaryParams{
		ProjectID: projectUUID,
		StartHash: lastCommitHash,
		EndHash:   commitHash,
	})
	if err == nil {
		summary = cached.SummaryRu
	} else {
		messages := make([]string, len(commits))
		for i, c := range commits {
			messages[i] = c.Message
		}

		summary, err = p.aiClient.SummarizeCommits(ctx, messages)
		if err != nil {
			return fmt.Errorf("failed to summarize commits: %w", err)
		}

		p.queries.CreateCommitsSummary(ctx, &db.CreateCommitsSummaryParams{
			ProjectID:   projectUUID,
			StartHash:   lastCommitHash,
			EndHash:     commitHash,
			SummaryRu:   summary,
			CommitCount: int32(len(commits)),
		})
	}

	buildUUID, err := uuid.Parse(buildID)
	if err != nil {
		return fmt.Errorf("invalid build ID: %w", err)
	}

	_, err = p.queries.GetBuildByID(ctx, buildUUID)
	if err != nil {
		return fmt.Errorf("build not found: %w", err)
	}

	// Загружаем артефакты из новой таблицы
	artifacts, err := p.artifactService.GetBuildArtifactsAsInterface(ctx, buildUUID)
	if err != nil {
		return fmt.Errorf("failed to get artifacts: %w", err)
	}

	// Собираем информацию об артефактах
	var artifactURLs []string
	var containerImages []string

	for _, artifact := range artifacts {
		if artMap, ok := artifact.(map[string]interface{}); ok {
			artifactType, _ := artMap["artifact_type"].(string)
			publicURL, _ := artMap["public_url"].(string)

			if artifactType == "file" && publicURL != "" {
				filename, _ := artMap["filename"].(string)
				artifactURLs = append(artifactURLs, fmt.Sprintf("%s - %s", filename, publicURL))
			} else if artifactType == "container_image" {
				imageName, _ := artMap["image_name"].(string)
				imageTag, _ := artMap["image_tag"].(string)
				imageDigest, _ := artMap["image_digest"].(string)
				// Если imageTag уже содержит полное имя образа (старый формат), используем его
				// Иначе формируем imageName:imageTag
				var imageStr string
				if imageTag == imageName || strings.Contains(imageTag, "/") {
					imageStr = imageTag
				} else {
					imageStr = imageName + ":" + imageTag
				}
				containerImages = append(containerImages, fmt.Sprintf("%s (%s)", imageStr, imageDigest[:12]))
			}
		}
	}

	webhookResults, err := p.notifier.SendWebhooks(ctx, &project, &branch, commitHash)
	if err != nil {
		return fmt.Errorf("failed to send webhooks: %w", err)
	}
	for _, result := range webhookResults {
		status := "webhook_success"
		message := fmt.Sprintf("Webhook: %s - HTTP %d", result.URL, result.StatusCode)
		if result.Error != nil {
			status = "webhook_failed"
			message = fmt.Sprintf("Webhook: %s - Error: %v", result.URL, result.Error)
		}
		p.queries.CreateBuildLog(ctx, &db.CreateBuildLogParams{
			BuildID:    buildUUID,
			ProjectID:  projectUUID,
			BranchID:   branchUUID,
			CommitHash: commitHash,
			Status:     status,
			LogMessage: message,
		})
	}

	// Отправляем уведомление с артефактами и образами
	if err := p.notifier.SendBuildNotificationWithArtifacts(ctx, &project, &branch, commitHash, authorName, summary, artifactURLs, containerImages); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	if err := p.executeSSHActions(ctx, &project, &branch, buildUUID, projectUUID, branchUUID, commitHash); err != nil {
		slog.Error("Failed to execute ssh actions", "error", err)
	}

	p.queries.UpdateBranchLastSuccessful(ctx, &db.UpdateBranchLastSuccessfulParams{
		ID:                   branchUUID,
		LastSuccessfulCommit: &commitHash,
		LastSuccessfulAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
	})

	return nil
}

func (p *Processor) executeSSHActions(ctx context.Context, project *db.Project, branch *db.Branch, buildUUID, projectUUID, branchUUID uuid.UUID, commitHash string) error {
	if p.sshExecutor == nil {
		return nil
	}

	var projectSettings map[string]interface{}
	if err := json.Unmarshal(project.Settings, &projectSettings); err != nil {
		return fmt.Errorf("parse project settings: %w", err)
	}

	var branchSettings map[string]interface{}
	if err := json.Unmarshal(branch.Settings, &branchSettings); err != nil {
		return fmt.Errorf("parse branch settings: %w", err)
	}

	var actions []map[string]interface{}
	if branchActions, ok := branchSettings["ssh_actions"].([]interface{}); ok && len(branchActions) > 0 {
		for _, a := range branchActions {
			if m, ok := a.(map[string]interface{}); ok {
				actions = append(actions, m)
			}
		}
	}
	if projectActions, ok := projectSettings["ssh_actions"].([]interface{}); ok && len(projectActions) > 0 {
		for _, a := range projectActions {
			if m, ok := a.(map[string]interface{}); ok {
				actions = append(actions, m)
			}
		}
	}

	for _, m := range actions {
		host, _ := m["host"].(string)
		username, _ := m["username"].(string)
		sshKeyID, _ := m["ssh_key_id"].(string)
		command, _ := m["command"].(string)
		if host == "" || username == "" || sshKeyID == "" || command == "" {
			continue
		}

		port := 22
		if p, ok := m["port"].(float64); ok {
			port = int(p)
		} else if p, ok := m["port"].(int); ok {
			port = p
		} else if ps, ok := m["port"].(string); ok {
			if pv, err := strconv.Atoi(ps); err == nil {
				port = pv
			}
		}

		action := &ssh.SSHAction{
			Host:     host,
			Port:     port,
			Username: username,
			SSHKeyID: sshKeyID,
			Command:  command,
		}
		result, err := p.sshExecutor.Execute(ctx, action)
		status := "ssh_success"
		var message string
		if err != nil {
			status = "ssh_failed"
			message = fmt.Sprintf("SSH: %s@%s\nCommand: %s\nError: %v", username, host, command, err)
			if result != nil && result.Stdout != "" {
				message += fmt.Sprintf("\nOutput:\n%s", result.Stdout)
			}
			slog.Error("SSH action failed", "host", host, "error", err)
		} else {
			stdout := ""
			if result != nil {
				stdout = result.Stdout
			}
			message = fmt.Sprintf("SSH: %s@%s\nCommand: %s\nOutput:\n%s", username, host, command, stdout)
			if result != nil && result.Stderr != "" {
				message += fmt.Sprintf("\nStderr:\n%s", result.Stderr)
			}
		}
		p.queries.CreateBuildLog(ctx, &db.CreateBuildLogParams{
			BuildID:    buildUUID,
			ProjectID:  projectUUID,
			BranchID:   branchUUID,
			CommitHash: commitHash,
			Status:     status,
			LogMessage: message,
		})
	}

	return nil
}

func (p *Processor) processRedeploy(ctx context.Context, task *Task) error {
	projectID := task.Data["project_id"].(string)
	branchID := task.Data["branch_id"].(string)
	commitHash := task.Data["commit_hash"].(string)
	buildID := task.Data["build_id"].(string)

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	branchUUID, err := uuid.Parse(branchID)
	if err != nil {
		return fmt.Errorf("invalid branch ID: %w", err)
	}

	buildUUID, err := uuid.Parse(buildID)
	if err != nil {
		return fmt.Errorf("invalid build ID: %w", err)
	}

	project, err := p.queries.GetProjectByID(ctx, projectUUID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	branch, err := p.queries.GetBranchByID(ctx, branchUUID)
	if err != nil {
		return fmt.Errorf("branch not found: %w", err)
	}

	webhookResults, err := p.notifier.SendWebhooks(ctx, &project, &branch, commitHash)
	if err != nil {
		return fmt.Errorf("failed to send webhooks: %w", err)
	}
	for _, result := range webhookResults {
		status := "webhook_success"
		message := fmt.Sprintf("Webhook: %s - HTTP %d", result.URL, result.StatusCode)
		if result.Error != nil {
			status = "webhook_failed"
			message = fmt.Sprintf("Webhook: %s - Error: %v", result.URL, result.Error)
		}
		p.queries.CreateBuildLog(ctx, &db.CreateBuildLogParams{
			BuildID:    buildUUID,
			ProjectID:  projectUUID,
			BranchID:   branchUUID,
			CommitHash: commitHash,
			Status:     status,
			LogMessage: message,
		})
	}

	if err := p.executeSSHActions(ctx, &project, &branch, buildUUID, projectUUID, branchUUID, commitHash); err != nil {
		slog.Error("Failed to execute ssh actions", "error", err)
	}

	return nil
}

func (p *Processor) processFailed(ctx context.Context, task *Task) error {
	projectID := task.Data["project_id"].(string)
	branchID := task.Data["branch_id"].(string)
	commitHash := task.Data["commit_hash"].(string)
	errorMessage := task.Data["error_message"].(string)

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	branchUUID, err := uuid.Parse(branchID)
	if err != nil {
		return fmt.Errorf("invalid branch ID: %w", err)
	}

	project, err := p.queries.GetProjectByID(ctx, projectUUID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	branch, err := p.queries.GetBranchByID(ctx, branchUUID)
	if err != nil {
		return fmt.Errorf("branch not found: %w", err)
	}

	webhookResults, err := p.notifier.SendWebhooks(ctx, &project, &branch, commitHash)
	if err != nil {
		return fmt.Errorf("failed to send webhooks: %w", err)
	}

	buildIDStr, _ := task.Data["build_id"].(string)
	if buildIDStr != "" {
		buildUUID, _ := uuid.Parse(buildIDStr)
		for _, result := range webhookResults {
			status := "webhook_success"
			message := fmt.Sprintf("Webhook: %s - HTTP %d", result.URL, result.StatusCode)
			if result.Error != nil {
				status = "webhook_failed"
				message = fmt.Sprintf("Webhook: %s - Error: %v", result.URL, result.Error)
			}
			p.queries.CreateBuildLog(ctx, &db.CreateBuildLogParams{
				BuildID:    buildUUID,
				ProjectID:  projectUUID,
				BranchID:   branchUUID,
				CommitHash: commitHash,
				Status:     status,
				LogMessage: message,
			})
		}
	}

	if err := p.notifier.SendFailedBuildNotification(ctx, &project, &branch, commitHash, errorMessage); err != nil {
		return fmt.Errorf("failed to send failed notification: %w", err)
	}

	return nil
}
