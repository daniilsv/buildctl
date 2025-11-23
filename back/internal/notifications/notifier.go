package notifications

import (
	"context"

	db "github.com/build-assistant/back/db/gen"
)

type Notifier interface {
	SendBuildNotification(ctx context.Context, project *db.Project, branch *db.Branch, commitHash, summary string, artifactURLs []string) error
	SendFailedBuildNotification(ctx context.Context, project *db.Project, branch *db.Branch, commitHash, errorMessage string) error
	SendWebhooks(ctx context.Context, project *db.Project, branch *db.Branch, commitHash string) error
	SendTestNotification(ctx context.Context, project *db.Project, branch *db.Branch) error
	SendTestWebhooks(ctx context.Context, project *db.Project, branch *db.Branch) error
}
