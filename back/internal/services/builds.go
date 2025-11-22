package services

import (
	"context"
	"fmt"
	"time"

	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/api/handlers/builds"
	"github.com/google/uuid"
)

type BuildService struct {
	queries *db.Queries
}

func NewBuildService(queries *db.Queries) *BuildService {
	return &BuildService{queries: queries}
}

func (s *BuildService) GetBuild(ctx context.Context, id string) (*builds.Build, error) {
	buildID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid build ID: %w", err)
	}

	dbBuild, err := s.queries.GetBuildByID(ctx, buildID)
	if err != nil {
		return nil, fmt.Errorf("build not found: %w", err)
	}

	project, err := s.queries.GetProjectByID(ctx, dbBuild.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	branch, err := s.queries.GetBranchByID(ctx, dbBuild.BranchID)
	if err != nil {
		return nil, fmt.Errorf("branch not found: %w", err)
	}

	logs, err := s.queries.GetBuildLogsByBuildID(ctx, buildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	build := toBuild(dbBuild, &project, &branch)
	build.Logs = make([]builds.Log, len(logs))
	for i, l := range logs {
		build.Logs[i] = toLog(l)
	}

	return build, nil
}

func (s *BuildService) ListBuilds(ctx context.Context, projectIDOrName, branchIDOrName *string, limit, offset int) ([]builds.Build, error) {
	var projectUUIDVal, branchUUIDVal uuid.UUID
	var hasProjectID, hasBranchID bool

	if projectIDOrName != nil && *projectIDOrName != "" {
		// Try to parse as UUID first, if fails treat as project name
		u, err := uuid.Parse(*projectIDOrName)
		if err != nil {
			// It's a project name, get project by name
			project, err := s.queries.GetProjectByName(ctx, *projectIDOrName)
			if err != nil {
				return nil, fmt.Errorf("project not found: %w", err)
			}
			projectUUIDVal = project.ID
		} else {
			// Verify it's a valid project UUID
			_, err := s.queries.GetProjectByID(ctx, u)
			if err != nil {
				return nil, fmt.Errorf("project not found: %w", err)
			}
			projectUUIDVal = u
		}
		hasProjectID = true
	}

	if branchIDOrName != nil && *branchIDOrName != "" {
		// Try to parse as UUID first, if fails treat as branch name
		u, err := uuid.Parse(*branchIDOrName)
		if err != nil {
			// It's a branch name, need project to find branch
			if !hasProjectID {
				return nil, fmt.Errorf("project_id or project_name required when using branch_name")
			}
			// Get project to find branch
			var project db.Project
			if projectUUIDVal != (uuid.UUID{}) {
				var err error
				project, err = s.queries.GetProjectByID(ctx, projectUUIDVal)
				if err != nil {
					return nil, fmt.Errorf("project not found: %w", err)
				}
			} else {
				return nil, fmt.Errorf("project_id or project_name required when using branch_name")
			}
			branch, err := s.queries.GetBranchByProjectAndName(ctx, &db.GetBranchByProjectAndNameParams{
				ProjectID: project.ID,
				Name:      *branchIDOrName,
			})
			if err != nil {
				return nil, fmt.Errorf("branch not found: %w", err)
			}
			branchUUIDVal = branch.ID
		} else {
			branchUUIDVal = u
		}
		hasBranchID = true
	}

	params := &db.ListBuildsParams{
		Column1: uuid.Nil, // Use Nil UUID to indicate "no filter"
		Column2: uuid.Nil, // Use Nil UUID to indicate "no filter"
		Limit:   int32(limit),
		Offset:  int32(offset),
	}

	if hasProjectID {
		params.Column1 = projectUUIDVal
	}
	if hasBranchID {
		params.Column2 = branchUUIDVal
	}

	dbBuilds, err := s.queries.ListBuilds(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list builds: %w", err)
	}

	result := make([]builds.Build, len(dbBuilds))
	for i, b := range dbBuilds {
		project, err := s.queries.GetProjectByID(ctx, b.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("project not found: %w", err)
		}

		branch, err := s.queries.GetBranchByID(ctx, b.BranchID)
		if err != nil {
			return nil, fmt.Errorf("branch not found: %w", err)
		}

		result[i] = *toBuild(b, &project, &branch)
	}

	return result, nil
}

func toBuild(b db.Build, project *db.Project, branch *db.Branch) *builds.Build {
	build := &builds.Build{
		ID:          b.ID.String(),
		ProjectID:   b.ProjectID.String(),
		ProjectName: project.Title,
		BranchID:    b.BranchID.String(),
		BranchName:   branch.Name,
		CommitHash:  b.CommitHash,
		Status:      b.Status,
		StartedAt:   b.StartedAt.Time.Format(time.RFC3339),
		CreatedAt:   b.CreatedAt.Time.Format(time.RFC3339),
	}

	if b.FinishedAt.Valid {
		build.FinishedAt = b.FinishedAt.Time.Format(time.RFC3339)
	}

	if b.CommitMessage != nil {
		build.CommitMessage = *b.CommitMessage
	}

	return build
}

func toLog(l db.BuildLog) builds.Log {
	log := builds.Log{
		ID:         l.ID.String(),
		Status:     l.Status,
		LogMessage: l.LogMessage,
		CreatedAt:  l.CreatedAt.Time.Format(time.RFC3339),
	}

	if l.ArtifactS3Key != nil {
		log.ArtifactS3Key = *l.ArtifactS3Key
	}
	if l.ArtifactUrl != nil {
		log.ArtifactURL = *l.ArtifactUrl
	}

	return log
}
