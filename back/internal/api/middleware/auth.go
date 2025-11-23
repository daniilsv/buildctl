package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"log/slog"

	"github.com/build-assistant/back/internal/auth"
)

type TokenValidator interface {
	ValidateToken(ctx context.Context, tokenHash string) error
}

type contextKey string

const (
	TokenContextKey contextKey = "token"
	UserContextKey  contextKey = "user"
)

type UserInfo struct {
	Subject string
	Email   string
}

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

func OIDCAuth(oidcService *auth.OIDCService, tokenCache *auth.TokenCache) func(http.Handler) http.Handler {
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

			accessToken := parts[1]

			// Check cache first
			tokenInfo, found := tokenCache.Get(accessToken)
			if !found {
				// Introspect token
				introspection, err := oidcService.IntrospectToken(r.Context(), accessToken)
				if err != nil {
					slog.Error("Token introspection failed", "error", err)
					http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
					return
				}

				// Calculate expiration time
				expiresAt := time.Now().Add(1 * time.Hour) // Default TTL
				if introspection.Exp > 0 {
					expiresAt = time.Unix(introspection.Exp, 0)
					// If token expires soon, use minimum TTL
					if expiresAt.Before(time.Now().Add(1 * time.Hour)) {
						expiresAt = time.Now().Add(1 * time.Hour)
					}
				}

				tokenInfo = &auth.TokenInfo{
					Subject:   introspection.Subject,
					Email:     introspection.Email,
					ExpiresAt: expiresAt,
				}

				// Cache the result
				tokenCache.Set(accessToken, tokenInfo)
			}

			// Add user info to context
			userInfo := UserInfo{
				Subject: tokenInfo.Subject,
				Email:   tokenInfo.Email,
			}
			ctx := context.WithValue(r.Context(), UserContextKey, userInfo)
			reqID := middleware.GetReqID(r.Context())
			slog.Info("OIDC auth", "req_id", reqID, "subject", userInfo.Subject)

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

func GetUserInfo(ctx context.Context) *UserInfo {
	if userInfo, ok := ctx.Value(UserContextKey).(UserInfo); ok {
		return &userInfo
	}
	return nil
}

