package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type IDTokenClaims struct {
	Subject string
	Email   string
	Name    string
}

type AuthService struct {
	oidcService  *OIDCService
	sessionStore SessionStore
}

func NewAuthService(oidcService *OIDCService, sessionStore SessionStore) *AuthService {
	return &AuthService{
		oidcService:  oidcService,
		sessionStore: sessionStore,
	}
}

func (s *AuthService) GenerateAuthURL() (string, string, error) {
	return s.oidcService.GenerateAuthURL()
}

func (s *AuthService) VerifyState(state string) bool {
	return s.oidcService.VerifyState(state)
}

func (s *AuthService) ExchangeCode(ctx context.Context, code string) (*IDTokenClaims, error) {
	idToken, err := s.oidcService.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	var claims struct {
		Subject string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &IDTokenClaims{
		Subject: claims.Subject,
		Email:   claims.Email,
		Name:    claims.Name,
	}, nil
}

func (s *AuthService) GetSession(r *http.Request) (*Session, error) {
	sessionID, err := GetSessionCookie(r)
	if err != nil {
		return nil, err
	}

	session, err := s.sessionStore.Get(sessionID)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AuthService) CreateSession(w http.ResponseWriter, userID, email string) error {
	sessionID, err := generateSessionID()
	if err != nil {
		return err
	}

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Email:     email,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.sessionStore.Set(sessionID, session); err != nil {
		return err
	}

	SetSessionCookie(w, sessionID)
	return nil
}

func (s *AuthService) ClearSession(w http.ResponseWriter, r *http.Request) error {
	sessionID, err := GetSessionCookie(r)
	if err != nil {
		return nil
	}

	s.sessionStore.Delete(sessionID)
	ClearSessionCookie(w)
	return nil
}
