package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/api/handlers/branches"
	"github.com/build-assistant/back/internal/workers"
)

type BranchService struct {
	queries    *db.Queries
	workerPool *workers.Pool
}

func NewBranchService(queries *db.Queries, workerPool *workers.Pool) *BranchService {
	return &BranchService{queries: queries, workerPool: workerPool}
}

func (s *BranchService) CreateBranch(ctx context.Context, projectName string, req branches.CreateBranchRequest) (*branches.Branch, error) {
	project, err := s.queries.GetProjectByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	settingsJSON, err := json.Marshal(req.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	dbBranch, err := s.queries.CreateBranch(ctx, &db.CreateBranchParams{
		ProjectID: project.ID,
		Name:      req.Name,
		Settings:  settingsJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	return toBranch(dbBranch), nil
}

func (s *BranchService) GetBranch(ctx context.Context, projectName, branchName string) (*branches.Branch, error) {
	project, err := s.queries.GetProjectByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	dbBranch, err := s.queries.GetBranchByProjectAndName(ctx, &db.GetBranchByProjectAndNameParams{
		ProjectID: project.ID,
		Name:      branchName,
	})
	if err != nil {
		return nil, fmt.Errorf("branch not found: %w", err)
	}

	return toBranch(dbBranch), nil
}

func (s *BranchService) ListBranches(ctx context.Context, projectName string) ([]branches.Branch, error) {
	project, err := s.queries.GetProjectByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	dbBranches, err := s.queries.ListBranches(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	result := make([]branches.Branch, len(dbBranches))
	for i, b := range dbBranches {
		result[i] = *toBranch(b)
	}

	return result, nil
}

func (s *BranchService) UpdateBranch(ctx context.Context, projectName, branchName string, req branches.UpdateBranchRequest) (*branches.Branch, error) {
	project, err := s.queries.GetProjectByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	branch, err := s.queries.GetBranchByProjectAndName(ctx, &db.GetBranchByProjectAndNameParams{
		ProjectID: project.ID,
		Name:      branchName,
	})
	if err != nil {
		return nil, fmt.Errorf("branch not found: %w", err)
	}

	settingsJSON, err := json.Marshal(req.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	dbBranch, err := s.queries.UpdateBranch(ctx, &db.UpdateBranchParams{
		ID:       branch.ID,
		Settings: settingsJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update branch: %w", err)
	}

	return toBranch(dbBranch), nil
}

func (s *BranchService) GetBuildsByBranch(ctx context.Context, projectName, branchName string) ([]branches.Build, error) {
	project, err := s.queries.GetProjectByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	branch, err := s.queries.GetBranchByProjectAndName(ctx, &db.GetBranchByProjectAndNameParams{
		ProjectID: project.ID,
		Name:      branchName,
	})
	if err != nil {
		return nil, fmt.Errorf("branch not found: %w", err)
	}

	dbBuilds, err := s.queries.GetBuildsByBranch(ctx, branch.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get builds: %w", err)
	}

	result := make([]branches.Build, len(dbBuilds))
	for i, b := range dbBuilds {
		result[i] = branches.Build{
			ID:         b.ID.String(),
			ProjectID:  b.ProjectID.String(),
			BranchID:   b.BranchID.String(),
			CommitHash: b.CommitHash,
			Status:     b.Status,
			StartedAt:  b.StartedAt.Time.Format(time.RFC3339),
			CreatedAt:  b.CreatedAt.Time.Format(time.RFC3339),
		}
		if b.FinishedAt.Valid {
			result[i].FinishedAt = b.FinishedAt.Time.Format(time.RFC3339)
		}
	}

	return result, nil
}

func toBranch(b db.Branch) *branches.Branch {
	var settings map[string]interface{}
	if len(b.Settings) > 0 {
		json.Unmarshal(b.Settings, &settings)
	}

	branch := &branches.Branch{
		ID:        b.ID.String(),
		ProjectID: b.ProjectID.String(),
		Name:      b.Name,
		Settings:  settings,
		CreatedAt: b.CreatedAt.Time.Format(time.RFC3339),
	}

	if b.LastSuccessfulCommit != nil {
		branch.LastSuccessfulCommit = *b.LastSuccessfulCommit
	}
	if b.LastSuccessfulAt.Valid {
		branch.LastSuccessfulAt = b.LastSuccessfulAt.Time.Format(time.RFC3339)
	}

	return branch
}

func (s *BranchService) GetProjectByNameDB(ctx context.Context, name string) (*db.Project, error) {
	dbProject, err := s.queries.GetProjectByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return &dbProject, nil
}

func (s *BranchService) GetBranchByProjectAndNameDB(ctx context.Context, projectID uuid.UUID, branchName string) (*db.Branch, error) {
	dbBranch, err := s.queries.GetBranchByProjectAndName(ctx, &db.GetBranchByProjectAndNameParams{
		ProjectID: projectID,
		Name:      branchName,
	})
	if err != nil {
		return nil, fmt.Errorf("branch not found: %w", err)
	}

	return &dbBranch, nil
}

func (s *BranchService) Redeploy(ctx context.Context, projectName, branchName string) error {
	project, err := s.queries.GetProjectByName(ctx, projectName)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	branch, err := s.queries.GetBranchByProjectAndName(ctx, &db.GetBranchByProjectAndNameParams{
		ProjectID: project.ID,
		Name:      branchName,
	})
	if err != nil {
		return fmt.Errorf("branch not found: %w", err)
	}

	if branch.LastSuccessfulCommit == nil || *branch.LastSuccessfulCommit == "" {
		return fmt.Errorf("no successful build for branch %s", branchName)
	}

	build, err := s.queries.GetBuildByProjectBranchCommit(ctx, &db.GetBuildByProjectBranchCommitParams{
		ProjectID:  project.ID,
		BranchID:   branch.ID,
		CommitHash: *branch.LastSuccessfulCommit,
	})
	if err != nil {
		return fmt.Errorf("last successful build not found: %w", err)
	}

	s.workerPool.Submit(&workers.Task{
		Type: workers.TaskTypeRedeploy,
		Data: map[string]interface{}{
			"project_id":  project.ID.String(),
			"branch_id":   branch.ID.String(),
			"commit_hash": *branch.LastSuccessfulCommit,
			"build_id":    build.ID.String(),
		},
	})

	return nil
}
