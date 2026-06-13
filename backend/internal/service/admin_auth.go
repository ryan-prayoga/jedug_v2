package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"jedug_backend/internal/repository"
)

const (
	AdminSessionCookieName   = "jedug_admin_session"
	AdminSessionCookiePath   = "/api/v1/admin"
	AdminSessionTTL          = 12 * time.Hour
	adminSessionOpTimeout    = 5 * time.Second
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

// hashSessionToken returns the SHA-256 hex digest of a raw session token.
// The raw token is sent to the browser cookie; only this hash is persisted.
func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// dbSessionStore is a DB-backed admin session store. Sessions survive restarts
// and enforce a single active session per admin via the repository.
type dbSessionStore struct {
	repo repository.AdminSessionRepository
	ttl  time.Duration
}

func newDBSessionStore(repo repository.AdminSessionRepository) *dbSessionStore {
	return &dbSessionStore{
		repo: repo,
		ttl:  AdminSessionTTL,
	}
}

// Create generates a fresh random token, persists its hash (replacing any prior
// session for the admin), and returns the RAW token for the cookie.
func (s *dbSessionStore) Create(username string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	ctx, cancel := context.WithTimeout(context.Background(), adminSessionOpTimeout)
	defer cancel()

	if err := s.repo.Create(ctx, hashSessionToken(token), username, time.Now().Add(s.ttl)); err != nil {
		return "", err
	}
	return token, nil
}

// Validate returns the session for a raw token if it is valid (unrevoked,
// unexpired), or nil otherwise.
func (s *dbSessionStore) Validate(token string) *AdminSession {
	if token == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), adminSessionOpTimeout)
	defer cancel()

	rec, err := s.repo.FindValid(ctx, hashSessionToken(token), time.Now())
	if err != nil || rec == nil {
		return nil
	}
	return &AdminSession{
		Username:  rec.Username,
		ExpiresAt: rec.ExpiresAt,
	}
}

// Revoke removes a session by raw token (logout).
func (s *dbSessionStore) Revoke(token string) {
	if token == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), adminSessionOpTimeout)
	defer cancel()

	_ = s.repo.Delete(ctx, hashSessionToken(token))
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
