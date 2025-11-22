package artifacts

import (
	"context"
	"encoding/json"
	"net/http"

	"log/slog"
)

type Service interface {
	PresignUpload(ctx context.Context, projectName, branchName, commitHash, filename string) (*PresignResponse, error)
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

type Handler struct {
	service        Service
	TokenValidator TokenValidator
}

func NewHandler(service Service, tokenValidator TokenValidator) *Handler {
	return &Handler{
		service:        service,
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
	json.NewEncoder(w).Encode(response)
}

