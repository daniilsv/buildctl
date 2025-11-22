package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
)

type TokenValidator interface {
	ValidateToken(ctx context.Context, tokenHash string) error
}

type contextKey string

const TokenContextKey contextKey = "token"

func TokenAuth(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			tokenHash := hashToken(token)

			if err := validator.ValidateToken(r.Context(), tokenHash); err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), TokenContextKey, tokenHash)
			reqID := middleware.GetReqID(r.Context())
			slog.Info("Token auth", "req_id", reqID, "token_hash", tokenHash[:8])

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func GetTokenHash(ctx context.Context) string {
	if tokenHash, ok := ctx.Value(TokenContextKey).(string); ok {
		return tokenHash
	}
	return ""
}

