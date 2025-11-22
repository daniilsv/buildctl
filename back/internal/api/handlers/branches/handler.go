package branches

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"log/slog"
)

type Service interface {
	CreateBranch(ctx context.Context, projectName string, req CreateBranchRequest) (*Branch, error)
	GetBranch(ctx context.Context, projectName, branchName string) (*Branch, error)
	ListBranches(ctx context.Context, projectName string) ([]Branch, error)
	UpdateBranch(ctx context.Context, projectName, branchName string, req UpdateBranchRequest) (*Branch, error)
	GetBuildsByBranch(ctx context.Context, projectName, branchName string) ([]Build, error)
}

type Build struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	BranchID    string `json:"branch_id"`
	CommitHash  string `json:"commit_hash"`
	Status      string `json:"status"`
	StartedAt   string `json:"started_at"`
	FinishedAt  string `json:"finished_at,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type UpdateBranchRequest struct {
	Settings map[string]interface{} `json:"settings"`
}

type Branch struct {
	ID                    string                 `json:"id"`
	ProjectID             string                 `json:"project_id"`
	Name                  string                 `json:"name"`
	LastSuccessfulCommit  string                 `json:"last_successful_commit,omitempty"`
	LastSuccessfulAt      string                 `json:"last_successful_at,omitempty"`
	Settings              map[string]interface{} `json:"settings"`
	CreatedAt             string                 `json:"created_at"`
}

type CreateBranchRequest struct {
	Name     string                 `json:"name"`
	Settings map[string]interface{} `json:"settings"`
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project_name")
	var req CreateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	branch, err := h.service.CreateBranch(r.Context(), projectName, req)
	if err != nil {
		slog.Error("Failed to create branch", "error", err)
		http.Error(w, "Failed to create branch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(branch)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project_name")
	branches, err := h.service.ListBranches(r.Context(), projectName)
	if err != nil {
		slog.Error("Failed to list branches", "error", err)
		http.Error(w, "Failed to list branches", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(branches)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project_name")
	branchName := chi.URLParam(r, "branch_name")
	branch, err := h.service.GetBranch(r.Context(), projectName, branchName)
	if err != nil {
		slog.Error("Failed to get branch", "error", err)
		http.Error(w, "Branch not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(branch)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project_name")
	branchName := chi.URLParam(r, "branch_name")
	var req UpdateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	branch, err := h.service.UpdateBranch(r.Context(), projectName, branchName, req)
	if err != nil {
		slog.Error("Failed to update branch", "error", err)
		http.Error(w, "Failed to update branch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(branch)
}

func (h *Handler) GetBuilds(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project_name")
	branchName := chi.URLParam(r, "branch_name")
	builds, err := h.service.GetBuildsByBranch(r.Context(), projectName, branchName)
	if err != nil {
		slog.Error("Failed to get builds", "error", err)
		http.Error(w, "Failed to get builds", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(builds)
}

