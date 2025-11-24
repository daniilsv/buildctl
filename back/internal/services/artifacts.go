package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/api/handlers/artifacts"
	"github.com/build-assistant/back/pkg/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ArtifactService struct {
	queries      *db.Queries
	s3Service    *s3.S3Service
	publicPrefix string
}

func NewArtifactService(queries *db.Queries, s3Service *s3.S3Service, publicPrefix string) *ArtifactService {
	return &ArtifactService{
		queries:      queries,
		s3Service:    s3Service,
		publicPrefix: publicPrefix,
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

type ArtifactWithURL struct {
	ID           string  `json:"id"`
	BuildID      string  `json:"build_id"`
	LogID        *string `json:"log_id,omitempty"`
	ProjectID    string  `json:"project_id"`
	BranchID     string  `json:"branch_id"`
	CommitHash   string  `json:"commit_hash"`
	Filename     string  `json:"filename"`
	S3Key        string  `json:"s3_key"`
	SizeBytes    int64   `json:"size_bytes"`
	ContentType  *string `json:"content_type,omitempty"`
	ArtifactType string  `json:"artifact_type"`
	ImageName    *string `json:"image_name,omitempty"`
	ImageTag     *string `json:"image_tag,omitempty"`
	ImageDigest  *string `json:"image_digest,omitempty"`
	CreatedAt    string  `json:"created_at"`
	DeletedAt    *string `json:"deleted_at,omitempty"`
	PublicURL    string  `json:"public_url"`
}

// convertArtifact converts db.Artifact to ArtifactWithURL
func convertArtifact(artifact db.Artifact, publicURL string) ArtifactWithURL {
	var logID *string
	if artifact.LogID.Valid {
		logIDStr := uuid.UUID(artifact.LogID.Bytes).String()
		logID = &logIDStr
	}

	var deletedAt *string
	if artifact.DeletedAt.Valid {
		deletedAtStr := artifact.DeletedAt.Time.Format("2006-01-02T15:04:05.999999Z07:00")
		deletedAt = &deletedAtStr
	}

	createdAt := ""
	if artifact.CreatedAt.Valid {
		createdAt = artifact.CreatedAt.Time.Format("2006-01-02T15:04:05.999999Z07:00")
	}

	return ArtifactWithURL{
		ID:           artifact.ID.String(),
		BuildID:      artifact.BuildID.String(),
		LogID:        logID,
		ProjectID:    artifact.ProjectID.String(),
		BranchID:     artifact.BranchID.String(),
		CommitHash:   artifact.CommitHash,
		Filename:     artifact.Filename,
		S3Key:        artifact.S3Key,
		SizeBytes:    artifact.SizeBytes,
		ContentType:  artifact.ContentType,
		ArtifactType: artifact.ArtifactType,
		ImageName:    artifact.ImageName,
		ImageTag:     artifact.ImageTag,
		ImageDigest:  artifact.ImageDigest,
		CreatedAt:    createdAt,
		DeletedAt:    deletedAt,
		PublicURL:    publicURL,
	}
}

func (s *ArtifactService) CreateArtifact(ctx context.Context, buildID uuid.UUID, logID *uuid.UUID, filename, s3Key string, sizeBytes int64, contentType *string) (any, error) {
	build, err := s.queries.GetBuildByID(ctx, buildID)
	if err != nil {
		return nil, fmt.Errorf("build not found: %w", err)
	}

	var logIDParam pgtype.UUID
	if logID != nil {
		logIDParam = pgtype.UUID{
			Bytes: *logID,
			Valid: true,
		}
	}

	artifact, err := s.queries.CreateArtifact(ctx, &db.CreateArtifactParams{
		BuildID:      buildID,
		LogID:        logIDParam,
		ProjectID:    build.ProjectID,
		BranchID:     build.BranchID,
		CommitHash:   build.CommitHash,
		Filename:     filename,
		S3Key:        s3Key,
		SizeBytes:    sizeBytes,
		ContentType:  contentType,
		ArtifactType: "file",
		ImageName:    nil,
		ImageTag:     nil,
		ImageDigest:  nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create artifact: %w", err)
	}

	result := convertArtifact(artifact, s.publicPrefix+s3Key)
	return &result, nil
}

func (s *ArtifactService) CreateContainerImage(ctx context.Context, buildID uuid.UUID, logID *uuid.UUID, imageName, imageTag, imageDigest string, filename *string, s3Key *string, sizeBytes *int64) (any, error) {
	build, err := s.queries.GetBuildByID(ctx, buildID)
	if err != nil {
		return nil, fmt.Errorf("build not found: %w", err)
	}

	var logIDParam pgtype.UUID
	if logID != nil {
		logIDParam = pgtype.UUID{
			Bytes: *logID,
			Valid: true,
		}
	}

	var filenameStr string
	if filename != nil {
		filenameStr = *filename
	} else {
		// Если imageTag уже содержит полное имя образа (старый формат с '/' или равен imageName),
		// используем его. Иначе формируем imageName:imageTag
		if imageTag == imageName || strings.Contains(imageTag, "/") {
			filenameStr = imageTag
		} else {
			filenameStr = imageName + ":" + imageTag
		}
	}

	s3KeyStr := ""
	if s3Key != nil {
		s3KeyStr = *s3Key
	}

	var sizeBytesVal int64
	if sizeBytes != nil {
		sizeBytesVal = *sizeBytes
	}

	artifact, err := s.queries.CreateArtifact(ctx, &db.CreateArtifactParams{
		BuildID:      buildID,
		LogID:        logIDParam,
		ProjectID:    build.ProjectID,
		BranchID:     build.BranchID,
		CommitHash:   build.CommitHash,
		Filename:     filenameStr,
		S3Key:        s3KeyStr,
		SizeBytes:    sizeBytesVal,
		ContentType:  nil,
		ArtifactType: "container_image",
		ImageName:    &imageName,
		ImageTag:     &imageTag,
		ImageDigest:  &imageDigest,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container image: %w", err)
	}

	publicURL := ""
	if s3Key != nil && *s3Key != "" {
		publicURL = s.publicPrefix + *s3Key
	}

	result := convertArtifact(artifact, publicURL)
	return &result, nil
}

func (s *ArtifactService) GetBuildArtifacts(ctx context.Context, buildID uuid.UUID) (any, error) {
	return s.getBuildArtifactsInternal(ctx, buildID)
}

func (s *ArtifactService) getBuildArtifactsInternal(ctx context.Context, buildID uuid.UUID) ([]ArtifactWithURL, error) {
	artifacts, err := s.queries.GetArtifactsByBuildID(ctx, buildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get artifacts: %w", err)
	}

	result := make([]ArtifactWithURL, len(artifacts))
	for i, artifact := range artifacts {
		publicURL := ""
		if artifact.S3Key != "" {
			publicURL = s.publicPrefix + artifact.S3Key
		}
		result[i] = convertArtifact(artifact, publicURL)
	}

	return result, nil
}

// GetBuildArtifactsAsInterface возвращает артефакты в виде []interface{} для использования в workers
func (s *ArtifactService) GetBuildArtifactsAsInterface(ctx context.Context, buildID uuid.UUID) ([]interface{}, error) {
	artifacts, err := s.getBuildArtifactsInternal(ctx, buildID)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(artifacts))
	for i, artifact := range artifacts {
		artMap := map[string]interface{}{
			"id":            artifact.ID,
			"artifact_type": artifact.ArtifactType,
			"filename":      artifact.Filename,
			"public_url":    artifact.PublicURL,
		}
		if artifact.ImageName != nil {
			artMap["image_name"] = *artifact.ImageName
		}
		if artifact.ImageTag != nil {
			artMap["image_tag"] = *artifact.ImageTag
		}
		if artifact.ImageDigest != nil {
			artMap["image_digest"] = *artifact.ImageDigest
		}
		result[i] = artMap
	}

	return result, nil
}

func (s *ArtifactService) DeleteArtifact(ctx context.Context, artifactID uuid.UUID) error {
	artifact, err := s.queries.GetArtifact(ctx, artifactID)
	if err != nil {
		return fmt.Errorf("artifact not found: %w", err)
	}

	// Попытка удаления из S3, игнорируем ошибки
	if artifact.S3Key != "" {
		_ = s.s3Service.DeleteObject(artifact.S3Key)
	}

	// Мягкое удаление в БД всегда успешно
	err = s.queries.SoftDeleteArtifact(ctx, artifactID)
	if err != nil {
		return fmt.Errorf("failed to soft delete artifact: %w", err)
	}

	return nil
}

func (s *ArtifactService) DeleteBuildArtifacts(ctx context.Context, buildID uuid.UUID) error {
	artifacts, err := s.queries.GetArtifactsByBuildID(ctx, buildID)
	if err != nil {
		return fmt.Errorf("failed to get artifacts: %w", err)
	}

	// Попытка удаления из S3, игнорируем ошибки
	for _, artifact := range artifacts {
		if artifact.S3Key != "" {
			_ = s.s3Service.DeleteObject(artifact.S3Key)
		}
	}

	// Мягкое удаление всех артефактов билда
	err = s.queries.SoftDeleteArtifactsByBuildID(ctx, buildID)
	if err != nil {
		return fmt.Errorf("failed to soft delete artifacts: %w", err)
	}

	return nil
}
