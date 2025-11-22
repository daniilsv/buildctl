package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/build-assistant/back/internal/auth"
)

type Service interface {
	GenerateAuthURL() (string, string, error)
	VerifyState(state string) bool
	ExchangeCode(ctx context.Context, code string) (*IDTokenClaims, error)
	GetSession(r *http.Request) (*auth.Session, error)
	CreateSession(w http.ResponseWriter, userID, email string) error
	ClearSession(w http.ResponseWriter, r *http.Request) error
}

type IDTokenClaims = auth.IDTokenClaims

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

	claims, err := h.service.ExchangeCode(r.Context(), code)
	if err != nil {
		slog.Error("Failed to exchange code", "error", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	if err := h.service.CreateSession(w, claims.Subject, claims.Email); err != nil {
		slog.Error("Failed to create session", "error", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.service.ClearSession(w, r); err != nil {
		slog.Error("Failed to clear session", "error", err)
	}
	http.Redirect(w, r, "/auth/login", http.StatusFound)
}

func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := h.service.GetSession(r)
		if err != nil {
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	session, err := h.service.GetSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"user_id": session.UserID,
		"email":   session.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
