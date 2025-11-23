package branches

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"log/slog"
	"github.com/google/uuid"
	"github.com/build-assistant/back/internal/notifications"
	db "github.com/build-assistant/back/db/gen"
)

type Service interface {
	CreateBranch(ctx context.Context, projectName string, req CreateBranchRequest) (*Branch, error)
	GetBranch(ctx context.Context, projectName, branchName string) (*Branch, error)
	ListBranches(ctx context.Context, projectName string) ([]Branch, error)
	UpdateBranch(ctx context.Context, projectName, branchName string, req UpdateBranchRequest) (*Branch, error)
	GetBuildsByBranch(ctx context.Context, projectName, branchName string) ([]Build, error)
	GetProjectByNameDB(ctx context.Context, name string) (*db.Project, error)
	GetBranchByProjectAndNameDB(ctx context.Context, projectID uuid.UUID, branchName string) (*db.Branch, error)
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
	service  Service
	notifier notifications.Notifier
}

func NewHandler(service Service, notifier notifications.Notifier) *Handler {
	return &Handler{service: service, notifier: notifier}
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

func (h *Handler) TestTelegramNotification(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project_name")
	branchName := chi.URLParam(r, "branch_name")

	dbProject, err := h.service.GetProjectByNameDB(r.Context(), projectName)
	if err != nil {
		slog.Error("Failed to get project", "error", err)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	dbBranch, err := h.service.GetBranchByProjectAndNameDB(r.Context(), dbProject.ID, branchName)
	if err != nil {
		slog.Error("Failed to get branch", "error", err)
		http.Error(w, "Branch not found", http.StatusNotFound)
		return
	}

	var req struct {
		ChatID   string  `json:"chat_id"`
		ThreadID *string `json:"thread_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ChatID == "" {
		http.Error(w, "chat_id is required", http.StatusBadRequest)
		return
	}

	if err := h.notifier.SendTestTelegramNotification(r.Context(), dbProject, dbBranch, req.ChatID, req.ThreadID); err != nil {
		slog.Error("Failed to send test notification", "error", err)
		http.Error(w, "Failed to send test notification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Test notification sent"})
}

func (h *Handler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project_name")
	branchName := chi.URLParam(r, "branch_name")

	dbProject, err := h.service.GetProjectByNameDB(r.Context(), projectName)
	if err != nil {
		slog.Error("Failed to get project", "error", err)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	dbBranch, err := h.service.GetBranchByProjectAndNameDB(r.Context(), dbProject.ID, branchName)
	if err != nil {
		slog.Error("Failed to get branch", "error", err)
		http.Error(w, "Branch not found", http.StatusNotFound)
		return
	}

	var req struct {
		WebhookURL string `json:"webhook_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.WebhookURL == "" {
		http.Error(w, "webhook_url is required", http.StatusBadRequest)
		return
	}

	if err := h.notifier.SendTestWebhook(r.Context(), dbProject, dbBranch, req.WebhookURL); err != nil {
		slog.Error("Failed to send test webhook", "error", err)
		http.Error(w, "Failed to send test webhook: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Test webhook sent"})
}

