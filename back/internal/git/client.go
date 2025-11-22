package git

import (
	"context"

	db "github.com/build-assistant/back/db/gen"
)

type Commit struct {
	Hash    string
	Message string
	Author  string
	Date    string
}

type Client interface {
	GetCommits(ctx context.Context, project *db.Project, branch *db.Branch, startHash, endHash string) ([]Commit, error)
	GetCommitByHash(ctx context.Context, project *db.Project, branch *db.Branch, commitHash string) (*Commit, error)
}
