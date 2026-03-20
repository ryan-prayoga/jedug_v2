package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const (
	AdminSessionCookieName   = "jedug_admin_session"
	AdminSessionCookiePath   = "/api/v1/admin"
	AdminSessionTTL          = 12 * time.Hour
	adminLoginWindow         = 15 * time.Minute
	adminLoginLockout        = 30 * time.Minute
	adminLoginMaxFailures    = 5
	adminLoginCleanupHorizon = 2 * time.Hour
)

// AdminSession represents an authenticated admin session.
type AdminSession struct {
	Username  string
	ExpiresAt time.Time
}

// SessionStore is a thread-safe in-memory session store.
type SessionStore struct {
	mu           sync.RWMutex
	sessions     map[string]*AdminSession
	tokenByAdmin map[string]string
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions:     make(map[string]*AdminSession),
		tokenByAdmin: make(map[string]string),
	}
}

func (s *SessionStore) Create(username string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpiredLocked(time.Now())
	if previous, ok := s.tokenByAdmin[username]; ok {
		delete(s.sessions, previous)
	}
	s.sessions[token] = &AdminSession{
		Username:  username,
		ExpiresAt: time.Now().Add(AdminSessionTTL),
	}
	s.tokenByAdmin[username] = token

	return token, nil
}

func (s *SessionStore) Validate(token string) *AdminSession {
	now := time.Now()

	s.mu.RLock()
	sess, ok := s.sessions[token]
	s.mu.RUnlock()

	if !ok {
		return nil
	}
	if now.After(sess.ExpiresAt) {
		s.Revoke(token)
		return nil
	}
	return sess
}

func (s *SessionStore) Revoke(token string) {
	if token == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[token]
	if !ok {
		return
	}
	delete(s.sessions, token)
	if current, ok := s.tokenByAdmin[sess.Username]; ok && current == token {
		delete(s.tokenByAdmin, sess.Username)
	}
}

func (s *SessionStore) cleanupExpiredLocked(now time.Time) {
	for token, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, token)
			if current, ok := s.tokenByAdmin[sess.Username]; ok && current == token {
				delete(s.tokenByAdmin, sess.Username)
			}
		}
	}
}

type adminLoginAttemptState struct {
	Failures    int
	WindowStart time.Time
	LastFailure time.Time
	LockedUntil time.Time
}

type adminLoginGuard struct {
	mu      sync.Mutex
	entries map[string]*adminLoginAttemptState
}

func newAdminLoginGuard() *adminLoginGuard {
	return &adminLoginGuard{
		entries: make(map[string]*adminLoginAttemptState),
	}
}

func (g *adminLoginGuard) Allow(key string) (bool, time.Duration) {
	if key == "" {
		return true, 0
	}

	now := time.Now()

	g.mu.Lock()
	defer g.mu.Unlock()

	g.cleanupLocked(now)
	state, ok := g.entries[key]
	if !ok {
		return true, 0
	}
	if state.LockedUntil.After(now) {
		return false, time.Until(state.LockedUntil)
	}
	return true, 0
}

func (g *adminLoginGuard) RecordFailure(key string) {
	if key == "" {
		return
	}

	now := time.Now()

	g.mu.Lock()
	defer g.mu.Unlock()

	g.cleanupLocked(now)
	state, ok := g.entries[key]
	if !ok || now.Sub(state.WindowStart) > adminLoginWindow {
		state = &adminLoginAttemptState{
			Failures:    0,
			WindowStart: now,
		}
		g.entries[key] = state
	}

	state.Failures++
	state.LastFailure = now
	if state.Failures >= adminLoginMaxFailures {
		state.LockedUntil = now.Add(adminLoginLockout)
	}
}

func (g *adminLoginGuard) Reset(key string) {
	if key == "" {
		return
	}

	g.mu.Lock()
	delete(g.entries, key)
	g.mu.Unlock()
}

func (g *adminLoginGuard) cleanupLocked(now time.Time) {
	for key, state := range g.entries {
		if state.LockedUntil.After(now) {
			continue
		}
		if now.Sub(state.LastFailure) > adminLoginCleanupHorizon {
			delete(g.entries, key)
		}
	}
}
