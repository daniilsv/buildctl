package auth

import (
	"context"
	"fmt"
)

type IDTokenClaims struct {
	Subject string
	Email   string
	Name    string
}

type AuthService struct {
	oidcService *OIDCService
}

func NewAuthService(oidcService *OIDCService, _ interface{}) *AuthService {
	return &AuthService{
		oidcService: oidcService,
	}
}

func (s *AuthService) GenerateAuthURL() (string, string, error) {
	return s.oidcService.GenerateAuthURL()
}

func (s *AuthService) VerifyState(state string) bool {
	return s.oidcService.VerifyState(state)
}

func (s *AuthService) ExchangeCode(ctx context.Context, code string) (string, *IDTokenClaims, error) {
	token, err := s.oidcService.ExchangeCodeForToken(ctx, code)
	if err != nil {
		return "", nil, err
	}

	idToken, err := s.oidcService.VerifyIDToken(ctx, token.IDToken)
	if err != nil {
		return "", nil, err
	}

	var claims struct {
		Subject string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return "", nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return token.AccessToken, &IDTokenClaims{
		Subject: claims.Subject,
		Email:   claims.Email,
		Name:    claims.Name,
	}, nil
}

