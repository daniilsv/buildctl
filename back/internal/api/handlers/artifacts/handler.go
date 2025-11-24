package artifacts

import (
	"context"
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

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
	PublicURL    string  `json:"public_url"`
}

type Service interface {
	PresignUpload(ctx context.Context, projectName, branchName, commitHash, filename string) (*PresignResponse, error)
	CreateArtifact(ctx context.Context, buildID uuid.UUID, logID *uuid.UUID, filename, s3Key string, sizeBytes int64, contentType *string) (any, error)
	CreateContainerImage(ctx context.Context, buildID uuid.UUID, logID *uuid.UUID, imageName, imageTag, imageDigest string, filename *string, s3Key *string, sizeBytes *int64) (any, error)
	GetBuildArtifacts(ctx context.Context, buildID uuid.UUID) (any, error)
	DeleteArtifact(ctx context.Context, artifactID uuid.UUID) error
	DeleteBuildArtifacts(ctx context.Context, buildID uuid.UUID) error
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, tokenHash string) error
}

type PresignRequest struct {
	ProjectName string `json:"project_name"`
	BranchName  string `json:"branch_name"`
	CommitHash  string `json:"commit_hash"`
	Filename    string `json:"filename"`
}

type PresignResponse struct {
	UploadURL string `json:"upload_url"`
	S3Key     string `json:"s3_key"`
}

type ConfirmUploadRequest struct {
	ProjectName string  `json:"project_name"`
	BranchName  string  `json:"branch_name"`
	CommitHash  string  `json:"commit_hash"`
	LogID       *string `json:"log_id,omitempty"`
	S3Key       string  `json:"s3_key"`
	Filename    string  `json:"filename"`
	SizeBytes   int64   `json:"size_bytes"`
	ContentType *string `json:"content_type,omitempty"`
}

type RegisterContainerRequest struct {
	ProjectName string  `json:"project_name"`
	BranchName  string  `json:"branch_name"`
	CommitHash  string  `json:"commit_hash"`
	LogID       *string `json:"log_id,omitempty"`
	ImageName   string  `json:"image_name"`
	ImageTag    string  `json:"image_tag"`
	ImageDigest string  `json:"image_digest"`
	Filename    *string `json:"filename,omitempty"`
	S3Key       *string `json:"s3_key,omitempty"`
	SizeBytes   *int64  `json:"size_bytes,omitempty"`
}

type BuildService interface {
	GetBuildByProjectBranchCommit(ctx context.Context, projectName, branchName, commitHash string) (map[string]interface{}, error)
}

type Handler struct {
	service        Service
	buildService   BuildService
	TokenValidator TokenValidator
}

func NewHandler(service Service, buildService BuildService, tokenValidator TokenValidator) *Handler {
	return &Handler{
		service:        service,
		buildService:   buildService,
		TokenValidator: tokenValidator,
	}
}

func (h *Handler) Presign(w http.ResponseWriter, r *http.Request) {
	var req PresignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.service.PresignUpload(r.Context(), req.ProjectName, req.BranchName, req.CommitHash, req.Filename)
	if err != nil {
		slog.Error("Failed to presign upload", "error", err)
		http.Error(w, "Failed to presign upload", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) ConfirmUpload(w http.ResponseWriter, r *http.Request) {
	var req ConfirmUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	build, err := h.buildService.GetBuildByProjectBranchCommit(r.Context(), req.ProjectName, req.BranchName, req.CommitHash)
	if err != nil {
		slog.Error("Failed to get build", "error", err)
		http.Error(w, "Build not found", http.StatusNotFound)
		return
	}

	buildIDStr := build["id"].(string)
	buildID, err := uuid.Parse(buildIDStr)
	if err != nil {
		http.Error(w, "Invalid build ID", http.StatusInternalServerError)
		return
	}

	var logID *uuid.UUID
	if req.LogID != nil {
		parsedLogID, err := uuid.Parse(*req.LogID)
		if err != nil {
			http.Error(w, "Invalid log ID", http.StatusBadRequest)
			return
		}
		logID = &parsedLogID
	}

	artifact, err := h.service.CreateArtifact(r.Context(), buildID, logID, req.Filename, req.S3Key, req.SizeBytes, req.ContentType)
	if err != nil {
		slog.Error("Failed to create artifact", "error", err)
		http.Error(w, "Failed to create artifact", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(artifact)
}

func (h *Handler) RegisterContainer(w http.ResponseWriter, r *http.Request) {
	var req RegisterContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	build, err := h.buildService.GetBuildByProjectBranchCommit(r.Context(), req.ProjectName, req.BranchName, req.CommitHash)
	if err != nil {
		slog.Error("Failed to get build", "error", err)
		http.Error(w, "Build not found", http.StatusNotFound)
		return
	}

	buildIDStr := build["id"].(string)
	buildID, err := uuid.Parse(buildIDStr)
	if err != nil {
		http.Error(w, "Invalid build ID", http.StatusInternalServerError)
		return
	}

	var logID *uuid.UUID
	if req.LogID != nil {
		parsedLogID, err := uuid.Parse(*req.LogID)
		if err != nil {
			http.Error(w, "Invalid log ID", http.StatusBadRequest)
			return
		}
		logID = &parsedLogID
	}

	artifact, err := h.service.CreateContainerImage(r.Context(), buildID, logID, req.ImageName, req.ImageTag, req.ImageDigest, req.Filename, req.S3Key, req.SizeBytes)
	if err != nil {
		slog.Error("Failed to create container image", "error", err)
		http.Error(w, "Failed to create container image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(artifact)
}

func (h *Handler) GetBuildArtifacts(w http.ResponseWriter, r *http.Request) {
	buildIDStr := chi.URLParam(r, "id")
	buildID, err := uuid.Parse(buildIDStr)
	if err != nil {
		http.Error(w, "Invalid build ID", http.StatusBadRequest)
		return
	}

	artifacts, err := h.service.GetBuildArtifacts(r.Context(), buildID)
	if err != nil {
		slog.Error("Failed to get artifacts", "error", err)
		http.Error(w, "Failed to get artifacts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(artifacts)
}

func (h *Handler) DeleteArtifact(w http.ResponseWriter, r *http.Request) {
	artifactIDStr := chi.URLParam(r, "id")
	artifactID, err := uuid.Parse(artifactIDStr)
	if err != nil {
		http.Error(w, "Invalid artifact ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteArtifact(r.Context(), artifactID)
	if err != nil {
		slog.Error("Failed to delete artifact", "error", err)
		http.Error(w, "Failed to delete artifact", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteBuildArtifacts(w http.ResponseWriter, r *http.Request) {
	buildIDStr := chi.URLParam(r, "id")
	buildID, err := uuid.Parse(buildIDStr)
	if err != nil {
		http.Error(w, "Invalid build ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteBuildArtifacts(r.Context(), buildID)
	if err != nil {
		slog.Error("Failed to delete build artifacts", "error", err)
		http.Error(w, "Failed to delete build artifacts", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
