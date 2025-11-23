package projects

import (
	"context"
	"encoding/json"
	"net/http"

	"log/slog"

	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/notifications"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Service interface {
	CreateProject(ctx context.Context, req CreateProjectRequest) (*Project, error)
	GetProject(ctx context.Context, name string) (*Project, error)
	ListProjects(ctx context.Context) ([]Project, error)
	UpdateProject(ctx context.Context, name string, req UpdateProjectRequest) (*Project, error)
	DeleteProject(ctx context.Context, name string) error
	GetProjectByNameDB(ctx context.Context, name string) (*db.Project, error)
}

type Project struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Title          string                 `json:"title"`
	RepositoryURL  string                 `json:"repository_url"`
	RepositoryType string                 `json:"repository_type"`
	Settings       map[string]interface{} `json:"settings"`
	CreatedAt      string                 `json:"created_at"`
}

type CreateProjectRequest struct {
	Name           string                 `json:"name"`
	Title          *string                `json:"title,omitempty"`
	RepositoryURL  string                 `json:"repository_url"`
	RepositoryType string                 `json:"repository_type"`
	AccessToken    string                 `json:"access_token"`
	Settings       map[string]interface{} `json:"settings"`
}

type UpdateProjectRequest struct {
	Name           *string                `json:"name,omitempty"`
	Title          *string                `json:"title,omitempty"`
	RepositoryURL  *string                `json:"repository_url,omitempty"`
	RepositoryType *string                `json:"repository_type,omitempty"`
	AccessToken    *string                `json:"access_token,omitempty"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
}

type Handler struct {
	service  Service
	notifier notifications.Notifier
}

func NewHandler(service Service, notifier notifications.Notifier) *Handler {
	return &Handler{service: service, notifier: notifier}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	project, err := h.service.CreateProject(r.Context(), req)
	if err != nil {
		slog.Error("Failed to create project", "error", err)
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(project)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	project, err := h.service.GetProject(r.Context(), name)
	if err != nil {
		slog.Error("Failed to get project", "error", err)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(project)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(r.Context())
	if err != nil {
		slog.Error("Failed to list projects", "error", err)
		http.Error(w, "Failed to list projects", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(projects)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	project, err := h.service.UpdateProject(r.Context(), name, req)
	if err != nil {
		slog.Error("Failed to update project", "error", err)
		http.Error(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(project)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := h.service.DeleteProject(r.Context(), name); err != nil {
		slog.Error("Failed to delete project", "error", err)
		http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) TestNotifications(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	dbProject, err := h.service.GetProjectByNameDB(r.Context(), name)
	if err != nil {
		slog.Error("Failed to get project", "error", err)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	branch := &db.Branch{
		ID:        uuid.New(),
		ProjectID: dbProject.ID,
		Name:      "test",
		Settings:  []byte(`{}`),
	}

	if err := h.notifier.SendTestNotification(r.Context(), dbProject, branch); err != nil {
		slog.Error("Failed to send test notification", "error", err)
		http.Error(w, "Failed to send test notification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Test notification sent"})
}

func (h *Handler) TestWebhooks(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	dbProject, err := h.service.GetProjectByNameDB(r.Context(), name)
	if err != nil {
		slog.Error("Failed to get project", "error", err)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	branch := &db.Branch{
		ID:        uuid.New(),
		ProjectID: dbProject.ID,
		Name:      "test",
		Settings:  []byte(`{}`),
	}

	if err := h.notifier.SendTestWebhooks(r.Context(), dbProject, branch); err != nil {
		slog.Error("Failed to send test webhooks", "error", err)
		http.Error(w, "Failed to send test webhooks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Test webhooks sent"})
}
