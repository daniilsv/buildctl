package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type OIDCService struct {
	provider   *oidc.Provider
	config     oauth2.Config
	verifier   *oidc.IDTokenVerifier
	stateStore map[string]time.Time
}

func NewOIDCService(cfg OIDCConfig) (*OIDCService, error) {
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("OIDC issuer is required")
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.ClientID,
	})

	return &OIDCService{
		provider:   provider,
		config:     oauth2Config,
		verifier:   verifier,
		stateStore: make(map[string]time.Time),
	}, nil
}

func (s *OIDCService) GenerateAuthURL() (string, string, error) {
	state, err := generateState()
	if err != nil {
		return "", "", err
	}

	s.stateStore[state] = time.Now().Add(10 * time.Minute)

	url := s.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return url, state, nil
}

func (s *OIDCService) VerifyState(state string) bool {
	expiry, exists := s.stateStore[state]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		delete(s.stateStore, state)
		return false
	}
	delete(s.stateStore, state)
	return true
}

func (s *OIDCService) ExchangeCode(ctx context.Context, code string) (*oidc.IDToken, error) {
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in response")
	}

	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	return idToken, nil
}

func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *OIDCService) CleanupExpiredStates() {
	now := time.Now()
	for state, expiry := range s.stateStore {
		if now.After(expiry) {
			delete(s.stateStore, state)
		}
	}
}
