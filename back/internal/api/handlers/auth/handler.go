package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/build-assistant/back/internal/api/middleware"
	authpkg "github.com/build-assistant/back/internal/auth"
)

type Service interface {
	GenerateAuthURL() (string, string, error)
	VerifyState(state string) bool
	ExchangeCode(ctx context.Context, code string) (string, *IDTokenClaims, error)
}

type IDTokenClaims = authpkg.IDTokenClaims

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	authURL, state, err := h.service.GenerateAuthURL()
	if err != nil {
		slog.Error("Failed to generate auth URL", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   600,
	})

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		http.Error(w, "Missing state or code", http.StatusBadRequest)
		return
	}

	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != state {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	if !h.service.VerifyState(state) {
		http.Error(w, "Invalid or expired state", http.StatusBadRequest)
		return
	}

	accessToken, claims, err := h.service.ExchangeCode(r.Context(), code)
	if err != nil {
		slog.Error("Failed to exchange code", "error", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"access_token": accessToken,
		"user_id":      claims.Subject,
		"email":        claims.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) UserInfo(w http.ResponseWriter, r *http.Request) {
	userInfo := middleware.GetUserInfo(r.Context())
	if userInfo == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"user_id": userInfo.Subject,
		"email":   userInfo.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
