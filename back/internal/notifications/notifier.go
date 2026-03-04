package notifications

import (
	"context"

	db "github.com/build-assistant/back/db/gen"
)

type WebhookResult struct {
	URL        string
	StatusCode int
	Error      error
}

type Notifier interface {
	SendBuildNotification(ctx context.Context, project *db.Project, branch *db.Branch, commitHash, summary string, artifactURLs []string) error
	SendBuildNotificationWithArtifacts(ctx context.Context, project *db.Project, branch *db.Branch, commitHash, authorName, summary string, artifactURLs []string, containerImages []string) error
	SendFailedBuildNotification(ctx context.Context, project *db.Project, branch *db.Branch, commitHash, errorMessage string) error
	SendWebhooks(ctx context.Context, project *db.Project, branch *db.Branch, commitHash string) ([]WebhookResult, error)
	SendTestTelegramNotification(ctx context.Context, project *db.Project, branch *db.Branch, chatID string, threadID *string) error
	SendTestWebhook(ctx context.Context, project *db.Project, branch *db.Branch, webhookURL string) error
}
