package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

type IntrospectionResponse struct {
	Active   bool   `json:"active"`
	Subject  string `json:"sub,omitempty"`
	Email    string `json:"email,omitempty"`
	Exp      int64  `json:"exp,omitempty"`
	Username string `json:"username,omitempty"`
}

type OIDCService struct {
	provider      *oidc.Provider
	config        oauth2.Config
	verifier      *oidc.IDTokenVerifier
	stateStore    map[string]time.Time
	introspectURL string
	clientID      string
	clientSecret  string
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

	// Get introspection endpoint from provider metadata
	var introspectURL string
	if err := provider.Claims(&struct {
		IntrospectionEndpoint *string `json:"introspection_endpoint"`
	}{}); err == nil {

		// Try to get from well-known endpoint
		introspectURL = cfg.Issuer + "/oauth/v2/introspect"
		if !strings.HasSuffix(cfg.Issuer, "/") {
			introspectURL = cfg.Issuer + "/oauth/v2/introspect"
		}
	}

	return &OIDCService{
		provider:      provider,
		config:        oauth2Config,
		verifier:      verifier,
		stateStore:    make(map[string]time.Time),
		introspectURL: introspectURL,
		clientID:      cfg.ClientID,
		clientSecret:  cfg.ClientSecret,
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

type TokenResponse struct {
	AccessToken string
	IDToken     string
}

func (s *OIDCService) ExchangeCodeForToken(ctx context.Context, code string) (*TokenResponse, error) {
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in response")
	}

	return &TokenResponse{
		AccessToken: token.AccessToken,
		IDToken:     rawIDToken,
	}, nil
}

func (s *OIDCService) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	return idToken, nil
}

func (s *OIDCService) ExchangeCode(ctx context.Context, code string) (*oidc.IDToken, error) {
	tokenResp, err := s.ExchangeCodeForToken(ctx, code)
	if err != nil {
		return nil, err
	}

	return s.VerifyIDToken(ctx, tokenResp.IDToken)
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

func (s *OIDCService) IntrospectToken(ctx context.Context, accessToken string) (*IntrospectionResponse, error) {
	if s.introspectURL == "" {
		// Try to discover introspection endpoint
		s.introspectURL = s.config.Endpoint.TokenURL
		if strings.Contains(s.introspectURL, "/token") {
			s.introspectURL = strings.Replace(s.introspectURL, "/token", "/introspect", 1)
		} else {
			s.introspectURL = strings.TrimSuffix(s.introspectURL, "/") + "/introspect"
		}
	}

	data := url.Values{}
	data.Set("token", accessToken)
	data.Set("token_type_hint", "access_token")

	req, err := http.NewRequestWithContext(ctx, "POST", s.introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.clientID, s.clientSecret)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("introspection failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result IntrospectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode introspection response: %w", err)
	}

	if !result.Active {
		return nil, fmt.Errorf("token is not active")
	}

	return &result, nil
}
