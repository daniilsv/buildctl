package services

import (
	"context"
	"fmt"
	"path/filepath"

	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/api/handlers/artifacts"
	"github.com/build-assistant/back/pkg/s3"
	"github.com/google/uuid"
)

type ArtifactService struct {
	queries   *db.Queries
	s3Service *s3.S3Service
}

func NewArtifactService(queries *db.Queries, s3Service *s3.S3Service) *ArtifactService {
	return &ArtifactService{
		queries:   queries,
		s3Service: s3Service,
	}
}

func (s *ArtifactService) PresignUpload(ctx context.Context, projectName, branchName, commitHash, filename string) (*artifacts.PresignResponse, error) {
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

	_, err = s.queries.GetBuildByProjectBranchCommit(ctx, &db.GetBuildByProjectBranchCommitParams{
		ProjectID:  project.ID,
		BranchID:   branch.ID,
		CommitHash: commitHash,
	})
	if err != nil {
		return nil, fmt.Errorf("build not found: %w", err)
	}

	folder := fmt.Sprintf("builds/%s", commitHash)
	presignedReq, s3Key, err := s.s3Service.PresignPut(folder, filepath.Base(filename), false)
	if err != nil {
		return nil, fmt.Errorf("failed to presign upload: %w", err)
	}

	return &artifacts.PresignResponse{
		UploadURL: presignedReq.URL,
		S3Key:     s3Key,
	}, nil
}

func (s *ArtifactService) ConfirmUpload(ctx context.Context, buildID, s3Key string) error {
	buildUUID, err := uuid.Parse(buildID)
	if err != nil {
		return fmt.Errorf("invalid build ID: %w", err)
	}

	_, err = s.queries.GetBuildByID(ctx, buildUUID)
	if err != nil {
		return fmt.Errorf("build not found: %w", err)
	}

	presignedGet, err := s.s3Service.PresignGet(s3Key)
	if err != nil {
		return fmt.Errorf("failed to presign get: %w", err)
	}

	logs, err := s.queries.GetBuildLogsByBuildID(ctx, buildUUID)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	if len(logs) > 0 {
		lastLog := logs[len(logs)-1]
		s.queries.UpdateBuildLogArtifact(ctx, &db.UpdateBuildLogArtifactParams{
			ID:            lastLog.ID,
			ArtifactS3Key: &s3Key,
			ArtifactUrl:   &presignedGet.URL,
		})
	}

	return nil
}
