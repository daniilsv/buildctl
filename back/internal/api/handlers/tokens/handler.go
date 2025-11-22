package tokens

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"log/slog"
)

type Service interface {
	CreateToken(ctx context.Context, name string, expiresAt *int64) (string, error)
	ListTokens(ctx context.Context) ([]Token, error)
	DeleteToken(ctx context.Context, id string) error
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, tokenHash string) error
}

type Token struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CreatedAt  string `json:"created_at"`
	LastUsedAt string `json:"last_used_at,omitempty"`
	ExpiresAt  string `json:"expires_at,omitempty"`
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string `json:"name"`
		ExpiresAt *int64 `json:"expires_at,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.service.CreateToken(r.Context(), req.Name, req.ExpiresAt)
	if err != nil {
		slog.Error("Failed to create token", "error", err)
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.service.ListTokens(r.Context())
	if err != nil {
		slog.Error("Failed to list tokens", "error", err)
		http.Error(w, "Failed to list tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing token ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteToken(r.Context(), id); err != nil {
		slog.Error("Failed to delete token", "error", err)
		http.Error(w, "Failed to delete token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

