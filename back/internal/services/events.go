package services

import (
	"context"
	"fmt"
	"time"

	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/api/handlers/events"
	"github.com/build-assistant/back/internal/git"
	"github.com/build-assistant/back/internal/workers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type EventService struct {
	queries    *db.Queries
	workerPool *workers.Pool
	gitClient  git.Client
}

func NewEventService(queries *db.Queries, workerPool *workers.Pool, gitClient git.Client) *EventService {
	return &EventService{
		queries:    queries,
		workerPool: workerPool,
		gitClient:  gitClient,
	}
}

func (s *EventService) HandleEvent(ctx context.Context, req events.EventRequest) error {
	project, err := s.queries.GetProjectByName(ctx, req.ProjectName)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	branch, err := s.queries.GetBranchByProjectAndName(ctx, &db.GetBranchByProjectAndNameParams{
		ProjectID: project.ID,
		Name:      req.Branch,
	})
	if err != nil {
		return fmt.Errorf("branch not found: %w", err)
	}

	var buildID uuid.UUID
	existingBuild, err := s.queries.GetBuildByCommitHash(ctx, req.CommitHash)
	if err == nil {
		buildID = existingBuild.ID
		s.queries.UpdateBuildStatus(ctx, &db.UpdateBuildStatusParams{
			ID:     buildID,
			Status: req.Status,
			FinishedAt: pgtype.Timestamp{
				Time:  time.Now(),
				Valid: req.Status == "success" || req.Status == "failed",
			},
		})
	} else {
		if req.Status == "started" {
			var commitMessage *string
			if s.gitClient != nil {
				commit, err := s.gitClient.GetCommitByHash(ctx, &project, &branch, req.CommitHash)
				if err == nil {
					commitMessage = &commit.Message
				}
			}

			newBuild, err := s.queries.CreateBuild(ctx, &db.CreateBuildParams{
				ProjectID:     project.ID,
				BranchID:      branch.ID,
				CommitHash:    req.CommitHash,
				CommitMessage: commitMessage,
				Status:        req.Status,
				StartedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
			})
			if err != nil {
				return fmt.Errorf("failed to create build: %w", err)
			}
			buildID = newBuild.ID
		} else {
			return fmt.Errorf("build not found for commit %s", req.CommitHash)
		}
	}

	_, err = s.queries.CreateBuildLog(ctx, &db.CreateBuildLogParams{
		BuildID:    buildID,
		ProjectID:  project.ID,
		BranchID:   branch.ID,
		CommitHash: req.CommitHash,
		Status:     req.Status,
		LogMessage: req.Log,
	})
	if err != nil {
		return fmt.Errorf("failed to create build log: %w", err)
	}

	if req.Status == "success" {
		s.workerPool.Submit(&workers.Task{
			Type: workers.TaskTypeProcessSuccess,
			Data: map[string]interface{}{
				"project_id":  project.ID.String(),
				"branch_id":   branch.ID.String(),
				"commit_hash": req.CommitHash,
				"build_id":    buildID.String(),
			},
		})
	} else if req.Status == "failed" {
		errorMessage := req.Log
		if errorMessage == "" {
			errorMessage = "Сборка завершилась с ошибкой"
		}
		s.workerPool.Submit(&workers.Task{
			Type: workers.TaskTypeProcessFailed,
			Data: map[string]interface{}{
				"project_id":   project.ID.String(),
				"branch_id":    branch.ID.String(),
				"commit_hash":  req.CommitHash,
				"build_id":     buildID.String(),
				"error_message": errorMessage,
			},
		})
	}

	return nil
}
