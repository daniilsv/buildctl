package projects

import (
	"context"
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5"
)

type Service interface {
	CreateProject(ctx context.Context, req CreateProjectRequest) (*Project, error)
	GetProject(ctx context.Context, name string) (*Project, error)
	ListProjects(ctx context.Context) ([]Project, error)
	UpdateProject(ctx context.Context, name string, req UpdateProjectRequest) (*Project, error)
	DeleteProject(ctx context.Context, name string) error
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
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
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
	json.NewEncoder(w).Encode(project)
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
	json.NewEncoder(w).Encode(project)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(r.Context())
	if err != nil {
		slog.Error("Failed to list projects", "error", err)
		http.Error(w, "Failed to list projects", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
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
	json.NewEncoder(w).Encode(project)
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
