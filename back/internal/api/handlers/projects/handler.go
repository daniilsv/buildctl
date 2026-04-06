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

func (h *Handler) TestTelegramNotification(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	dbProject, err := h.service.GetProjectByNameDB(r.Context(), name)
	if err != nil {
		slog.Error("Failed to get project", "error", err)
		http.Error(w, "Project not found", http.StatusNotFound)
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

	branch := &db.Branch{
		ID:        uuid.New(),
		ProjectID: dbProject.ID,
		Name:      "test",
		Settings:  []byte(`{}`),
	}

	if err := h.notifier.SendTestTelegramNotification(r.Context(), dbProject, branch, req.ChatID, req.ThreadID); err != nil {
		slog.Error("Failed to send test notification", "error", err)
		http.Error(w, "Failed to send test notification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Test notification sent"})
}

func (h *Handler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	dbProject, err := h.service.GetProjectByNameDB(r.Context(), name)
	if err != nil {
		slog.Error("Failed to get project", "error", err)
		http.Error(w, "Project not found", http.StatusNotFound)
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

	branch := &db.Branch{
		ID:        uuid.New(),
		ProjectID: dbProject.ID,
		Name:      "test",
		Settings:  []byte(`{}`),
	}

	if err := h.notifier.SendTestWebhook(r.Context(), dbProject, branch, req.WebhookURL); err != nil {
		slog.Error("Failed to send test webhook", "error", err)
		http.Error(w, "Failed to send test webhook: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Test webhook sent"})
}

func (h *Handler) TestB24Notification(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	dbProject, err := h.service.GetProjectByNameDB(r.Context(), name)
	if err != nil {
		slog.Error("Failed to get project", "error", err)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	var req struct {
		TypeKey string `json:"type_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TypeKey == "" {
		http.Error(w, "type_key is required", http.StatusBadRequest)
		return
	}

	branch := &db.Branch{
		ID:        uuid.New(),
		ProjectID: dbProject.ID,
		Name:      "test",
		Settings:  []byte(`{}`),
	}

	if err := h.notifier.SendTestB24Notification(r.Context(), dbProject, branch, req.TypeKey); err != nil {
		slog.Error("Failed to send test B24 notification", "error", err)
		http.Error(w, "Failed to send test B24 notification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Test B24 notification sent"})
}
