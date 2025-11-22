package ai

import (
	"context"
)

type Client interface {
	SummarizeCommits(ctx context.Context, messages []string) (string, error)
}
