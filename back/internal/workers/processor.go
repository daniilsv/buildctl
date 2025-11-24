package workers

import (
	"context"
	"fmt"
	"strings"
	"time"

	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/ai"
	"github.com/build-assistant/back/internal/git"
	"github.com/build-assistant/back/internal/notifications"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
}

func NewProcessor(queries *db.Queries, gitClient git.Client, aiClient ai.Client, notifier notifications.Notifier, artifactService ArtifactService) *Processor {
	return &Processor{
		queries:         queries,
		gitClient:       gitClient,
		aiClient:        aiClient,
		notifier:        notifier,
		artifactService: artifactService,
	}
}

func (p *Processor) ProcessTask(ctx context.Context, task *Task) error {
	switch task.Type {
	case TaskTypeProcessSuccess:
		return p.processSuccess(ctx, task)
	case TaskTypeProcessFailed:
		return p.processFailed(ctx, task)
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

	// Отправляем уведомление с артефактами и образами
	if err := p.notifier.SendBuildNotificationWithArtifacts(ctx, &project, &branch, commitHash, summary, artifactURLs, containerImages); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	if err := p.notifier.SendWebhooks(ctx, &project, &branch, commitHash); err != nil {
		return fmt.Errorf("failed to send webhooks: %w", err)
	}

	p.queries.UpdateBranchLastSuccessful(ctx, &db.UpdateBranchLastSuccessfulParams{
		ID:                   branchUUID,
		LastSuccessfulCommit: &commitHash,
		LastSuccessfulAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
	})

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

	if err := p.notifier.SendFailedBuildNotification(ctx, &project, &branch, commitHash, errorMessage); err != nil {
		return fmt.Errorf("failed to send failed notification: %w", err)
	}

	if err := p.notifier.SendWebhooks(ctx, &project, &branch, commitHash); err != nil {
		return fmt.Errorf("failed to send webhooks: %w", err)
	}

	return nil
}
