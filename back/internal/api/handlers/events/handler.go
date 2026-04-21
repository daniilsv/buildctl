package events

import (
	"context"
	"encoding/json"
	"net/http"

	"log/slog"
)

type Service interface {
	HandleEvent(ctx context.Context, req EventRequest) error
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, tokenHash string) error
}

type EventRequest struct {
	ProjectName string `json:"project_name"`
	CommitHash  string `json:"commit_hash"`
	Branch      string `json:"branch"`
	Status      string `json:"status"`
	Log         string `json:"log"`
	BuildNumber string `json:"build_number,omitempty"`
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

func (h *Handler) HandleEvent(w http.ResponseWriter, r *http.Request) {
	var req EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.HandleEvent(r.Context(), req); err != nil {
		slog.Error("Failed to handle event", "error", err)
		http.Error(w, "Failed to handle event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
