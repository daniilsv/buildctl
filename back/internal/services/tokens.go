package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/api/handlers/tokens"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type TokenService struct {
	queries *db.Queries
}

func NewTokenService(queries *db.Queries) *TokenService {
	return &TokenService{queries: queries}
}

func (s *TokenService) ValidateToken(ctx context.Context, tokenHash string) error {
	token, err := s.queries.GetAccessTokenByHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("token not found")
	}

	if token.ExpiresAt.Valid && time.Now().After(token.ExpiresAt.Time) {
		return fmt.Errorf("token expired")
	}

	s.queries.UpdateAccessTokenLastUsed(ctx, token.ID)

	return nil
}

func (s *TokenService) CreateToken(ctx context.Context, name string, expiresAt *int64) (string, error) {
	token := generateToken()
	tokenHash := hashToken(token)

	var expiresAtTimestamp pgtype.Timestamp
	if expiresAt != nil {
		expiresAtTimestamp = pgtype.Timestamp{Time: time.Unix(*expiresAt, 0), Valid: true}
	} else {
		expiresAtTimestamp = pgtype.Timestamp{Valid: false}
	}

	_, err := s.queries.CreateAccessToken(ctx, &db.CreateAccessTokenParams{
		Name:      name,
		TokenHash: tokenHash,
		ExpiresAt: expiresAtTimestamp,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return token, nil
}

func (s *TokenService) ListTokens(ctx context.Context) ([]tokens.Token, error) {
	tokenList, err := s.queries.ListAccessTokens(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]tokens.Token, len(tokenList))
	for i, t := range tokenList {
		result[i] = tokens.Token{
			ID:         t.ID.String(),
			Name:       t.Name,
			CreatedAt:  t.CreatedAt.Time.Format(time.RFC3339),
			LastUsedAt: formatTime(&t.LastUsedAt.Time),
			ExpiresAt:  formatTime(&t.ExpiresAt.Time),
		}
	}

	return result, nil
}

func (s *TokenService) DeleteToken(ctx context.Context, id string) error {
	tokenID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid token ID: %w", err)
	}

	return s.queries.DeleteAccessToken(ctx, tokenID)
}

func generateToken() string {
	return uuid.New().String() + "-" + uuid.New().String()
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
