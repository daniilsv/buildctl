package ssh_keys

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"log/slog"
)

type Service interface {
	Create(ctx context.Context, name string, privateKey string) (*SSHKey, error)
	List(ctx context.Context) ([]SSHKey, error)
	Delete(ctx context.Context, id string) error
}

type SSHKey struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	CreatedAt   string `json:"created_at"`
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string `json:"name"`
		PrivateKey string `json:"private_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.PrivateKey == "" {
		http.Error(w, "name and private_key are required", http.StatusBadRequest)
		return
	}

	key, err := h.service.Create(r.Context(), req.Name, req.PrivateKey)
	if err != nil {
		slog.Error("Failed to create ssh key", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(key)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	keys, err := h.service.List(r.Context())
	if err != nil {
		slog.Error("Failed to list ssh keys", "error", err)
		http.Error(w, "Failed to list ssh keys", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing ssh key ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		slog.Error("Failed to delete ssh key", "error", err)
		http.Error(w, "Failed to delete ssh key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
