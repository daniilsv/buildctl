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

// ArtifactLink describes a single file artifact: a human-readable name and the
// URL it can be downloaded from. Keeping the two fields separate lets each
// notifier render them appropriately (plain text for Telegram, a proper BB-code
// link for B24) instead of pre-joining them into one ambiguous string.
type ArtifactLink struct {
	Name string
	URL  string
}

type Notifier interface {
	SendBuildNotification(ctx context.Context, project *db.Project, branch *db.Branch, commitHash, summary string, artifacts []ArtifactLink) error
	SendBuildNotificationWithArtifacts(ctx context.Context, project *db.Project, branch *db.Branch, commitHash, authorName, summary string, artifacts []ArtifactLink, containerImages []string, buildNumber string) error
	SendFailedBuildNotification(ctx context.Context, project *db.Project, branch *db.Branch, commitHash, errorMessage, buildNumber string) error
	SendWebhooks(ctx context.Context, project *db.Project, branch *db.Branch, commitHash string) ([]WebhookResult, error)
	SendTestTelegramNotification(ctx context.Context, project *db.Project, branch *db.Branch, chatID string, threadID *string) error
	SendTestWebhook(ctx context.Context, project *db.Project, branch *db.Branch, webhookURL string) error
	SendTestB24Notification(ctx context.Context, project *db.Project, branch *db.Branch, typeKey string) error
}
