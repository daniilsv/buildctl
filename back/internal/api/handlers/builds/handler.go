package builds

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/go-chi/chi/v5"
)

type Service interface {
	GetBuild(ctx context.Context, id string) (*Build, error)
	ListBuilds(ctx context.Context, projectID, branchID *string, limit, offset int) ([]Build, error)
	GetBuildByProjectBranchCommit(ctx context.Context, projectName, branchName, commitHash string) (map[string]interface{}, error)
}

type Build struct {
	ID            string `json:"id"`
	ProjectID     string `json:"project_id"`
	ProjectName   string `json:"project_name,omitempty"`
	BranchID      string `json:"branch_id"`
	BranchName    string `json:"branch_name,omitempty"`
	CommitHash    string `json:"commit_hash"`
	CommitMessage string `json:"commit_message,omitempty"`
	Status        string `json:"status"`
	StartedAt     string `json:"started_at"`
	FinishedAt    string `json:"finished_at,omitempty"`
	CreatedAt     string `json:"created_at"`
	Logs          []Log  `json:"logs,omitempty"`
}

type Log struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	LogMessage    string `json:"log_message"`
	ArtifactS3Key string `json:"artifact_s3_key,omitempty"`
	ArtifactURL   string `json:"artifact_url,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	build, err := h.service.GetBuild(r.Context(), id)
	if err != nil {
		slog.Error("Failed to get build", "error", err)
		http.Error(w, "Build not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(build)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	projectName := r.URL.Query().Get("project_name")
	branchID := r.URL.Query().Get("branch_id")
	branchName := r.URL.Query().Get("branch_name")

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	var projectIDPtr, branchIDPtr *string
	if projectName != "" {
		projectIDPtr = &projectName
	} else if projectID != "" {
		projectIDPtr = &projectID
	}
	if branchName != "" {
		branchIDPtr = &branchName
	} else if branchID != "" {
		branchIDPtr = &branchID
	}

	builds, err := h.service.ListBuilds(r.Context(), projectIDPtr, branchIDPtr, limit, offset)
	if err != nil {
		slog.Error("Failed to list builds", "error", err)
		http.Error(w, "Failed to list builds", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(builds)
}
