package workers

import (
	"context"
	"fmt"
	"time"

	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/ai"
	"github.com/build-assistant/back/internal/git"
	"github.com/build-assistant/back/internal/notifications"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Processor struct {
	queries   *db.Queries
	gitClient git.Client
	aiClient  ai.Client
	notifier  notifications.Notifier
}

func NewProcessor(queries *db.Queries, gitClient git.Client, aiClient ai.Client, notifier notifications.Notifier) *Processor {
	return &Processor{
		queries:   queries,
		gitClient: gitClient,
		aiClient:  aiClient,
		notifier:  notifier,
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

	logs, err := p.queries.GetBuildLogsByBuildID(ctx, buildUUID)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	var artifactURLs []string
	for _, log := range logs {
		if log.ArtifactUrl != nil {
			artifactURLs = append(artifactURLs, *log.ArtifactUrl)
		}
	}

	if err := p.notifier.SendBuildNotification(ctx, &project, &branch, commitHash, summary, artifactURLs); err != nil {
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
