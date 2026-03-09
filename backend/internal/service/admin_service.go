package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

const sessionTTL = 24 * time.Hour

// AdminSession represents an authenticated admin session.
type AdminSession struct {
	Username  string
	ExpiresAt time.Time
}

// SessionStore is a thread-safe in-memory session store.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*AdminSession
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]*AdminSession)}
}

func (s *SessionStore) Create(username string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	s.mu.Lock()
	s.sessions[token] = &AdminSession{
		Username:  username,
		ExpiresAt: time.Now().Add(sessionTTL),
	}
	s.mu.Unlock()

	return token, nil
}

func (s *SessionStore) Validate(token string) *AdminSession {
	s.mu.RLock()
	sess, ok := s.sessions[token]
	s.mu.RUnlock()

	if !ok || time.Now().After(sess.ExpiresAt) {
		if ok {
			s.mu.Lock()
			delete(s.sessions, token)
			s.mu.Unlock()
		}
		return nil
	}
	return sess
}

// AdminService handles admin authentication and moderation operations.
type AdminService interface {
	Login(username, password string) (string, error)
	ValidateSession(token string) *AdminSession
	ListIssues(ctx context.Context, limit, offset int, status *string) ([]*domain.AdminIssue, error)
	GetIssueDetail(ctx context.Context, id uuid.UUID) (*domain.AdminIssueDetail, error)
	HideIssue(ctx context.Context, id uuid.UUID, username string, reason *string) error
	UnhideIssue(ctx context.Context, id uuid.UUID, username string, reason *string) error
	FixIssue(ctx context.Context, id uuid.UUID, username string, reason *string) error
	RejectIssue(ctx context.Context, id uuid.UUID, username string, reason *string) error
	BanDevice(ctx context.Context, id uuid.UUID, username string, reason *string) error
}

type adminService struct {
	adminUsername string
	adminPassword string
	sessions      *SessionStore
	repo          repository.AdminRepository
}

func NewAdminService(username, password string, repo repository.AdminRepository) AdminService {
	return &adminService{
		adminUsername: username,
		adminPassword: password,
		sessions:      NewSessionStore(),
		repo:          repo,
	}
}

func (s *adminService) Login(username, password string) (string, error) {
	usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(s.adminUsername)) == 1
	passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(s.adminPassword)) == 1

	if !usernameMatch || !passwordMatch {
		return "", ErrInvalidCredentials
	}

	return s.sessions.Create(username)
}

func (s *adminService) ValidateSession(token string) *AdminSession {
	return s.sessions.Validate(token)
}

func (s *adminService) ListIssues(ctx context.Context, limit, offset int, status *string) ([]*domain.AdminIssue, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListIssues(ctx, limit, offset, status)
}

func (s *adminService) GetIssueDetail(ctx context.Context, id uuid.UUID) (*domain.AdminIssueDetail, error) {
	return s.repo.FindIssueByIDWithDetail(ctx, id)
}

func (s *adminService) HideIssue(ctx context.Context, id uuid.UUID, username string, reason *string) error {
	if err := s.repo.UpdateIssueHidden(ctx, id, true, reason); err != nil {
		return err
	}
	return s.repo.CreateModerationAction(ctx, "hide_issue", "issue", id, username, reason)
}

func (s *adminService) UnhideIssue(ctx context.Context, id uuid.UUID, username string, reason *string) error {
	if err := s.repo.UpdateIssueHidden(ctx, id, false, nil); err != nil {
		return err
	}
	return s.repo.CreateModerationAction(ctx, "unhide_issue", "issue", id, username, reason)
}

func (s *adminService) FixIssue(ctx context.Context, id uuid.UUID, username string, reason *string) error {
	if err := s.repo.UpdateIssueStatus(ctx, id, "fixed"); err != nil {
		return err
	}
	return s.repo.CreateModerationAction(ctx, "mark_fixed", "issue", id, username, reason)
}

func (s *adminService) RejectIssue(ctx context.Context, id uuid.UUID, username string, reason *string) error {
	if err := s.repo.UpdateIssueStatus(ctx, id, "rejected"); err != nil {
		return err
	}
	return s.repo.CreateModerationAction(ctx, "reject_issue", "issue", id, username, reason)
}

func (s *adminService) BanDevice(ctx context.Context, id uuid.UUID, username string, reason *string) error {
	if err := s.repo.BanDevice(ctx, id, reason); err != nil {
		return err
	}
	return s.repo.CreateModerationAction(ctx, "ban_device", "device", id, username, reason)
}

// GetSessionStore returns the session store for use by the auth middleware.
func (s *adminService) GetSessionStore() *SessionStore {
	return s.sessions
}
