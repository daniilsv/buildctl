package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

const (
	sessionCookieName = "build_assistant_session"
	sessionDuration   = 24 * time.Hour
)

type SessionStore interface {
	Get(sessionID string) (*Session, error)
	Set(sessionID string, session *Session) error
	Delete(sessionID string) error
}

type InMemorySessionStore struct {
	sessions map[string]*Session
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *InMemorySessionStore) Get(sessionID string) (*Session, error) {
	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, sessionID)
		return nil, fmt.Errorf("session expired")
	}
	return session, nil
}

func (s *InMemorySessionStore) Set(sessionID string, session *Session) error {
	s.sessions[sessionID] = session
	return nil
}

func (s *InMemorySessionStore) Delete(sessionID string) error {
	delete(s.sessions, sessionID)
	return nil
}

func (s *InMemorySessionStore) CleanupExpired() {
	now := time.Now()
	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}

type Session struct {
	ID        string
	UserID    string
	Email     string
	ExpiresAt time.Time
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func SetSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionDuration.Seconds()),
	}
	http.SetCookie(w, cookie)
}

func GetSessionCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}
