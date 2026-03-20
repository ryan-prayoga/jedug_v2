package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrModerationTargetNotFound = repository.ErrModerationTargetNotFound
var ErrAdminLoginThrottled = errors.New("admin login throttled")

const moderationAuditTimeout = 5 * time.Second

type AdminLoginThrottleError struct {
	RetryAfter time.Duration
}

func (e *AdminLoginThrottleError) Error() string {
	return ErrAdminLoginThrottled.Error()
}

// AdminService handles admin authentication and moderation operations.
type AdminService interface {
	Login(username, password, fingerprint string) (string, error)
	ValidateSession(token string) *AdminSession
	RevokeSession(token string)
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
	loginGuard    *adminLoginGuard
	repo          repository.AdminRepository
}

func NewAdminService(username, password string, repo repository.AdminRepository) AdminService {
	return &adminService{
		adminUsername: username,
		adminPassword: password,
		sessions:      NewSessionStore(),
		loginGuard:    newAdminLoginGuard(),
		repo:          repo,
	}
}

func (s *adminService) Login(username, password, fingerprint string) (string, error) {
	if allowed, retryAfter := s.loginGuard.Allow(fingerprint); !allowed {
		return "", &AdminLoginThrottleError{RetryAfter: retryAfter}
	}

	usernameMatch := constantTimeSecretMatch(username, s.adminUsername)
	passwordMatch := constantTimeSecretMatch(password, s.adminPassword)

	if !usernameMatch || !passwordMatch {
		s.loginGuard.RecordFailure(fingerprint)
		return "", ErrInvalidCredentials
	}

	s.loginGuard.Reset(fingerprint)
	return s.sessions.Create(username)
}

func (s *adminService) ValidateSession(token string) *AdminSession {
	return s.sessions.Validate(token)
}

func (s *adminService) RevokeSession(token string) {
	s.sessions.Revoke(token)
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
	s.recordModerationAction(ctx, "hide_issue", "issue", id, username, reason)
	return nil
}

func (s *adminService) UnhideIssue(ctx context.Context, id uuid.UUID, username string, reason *string) error {
	if err := s.repo.UpdateIssueHidden(ctx, id, false, nil); err != nil {
		return err
	}
	s.recordModerationAction(ctx, "unhide_issue", "issue", id, username, reason)
	return nil
}

func (s *adminService) FixIssue(ctx context.Context, id uuid.UUID, username string, reason *string) error {
	result, err := s.repo.ModerateIssueStatus(ctx, id, "fixed", 1)
	if err != nil {
		return err
	}
	if result.StatusChanged {
		s.publishStatusUpdated(ctx, id, result.PreviousStatus, "fixed")
	}
	s.recordModerationAction(ctx, "mark_fixed", "issue", id, username, reason)
	return nil
}

func (s *adminService) RejectIssue(ctx context.Context, id uuid.UUID, username string, reason *string) error {
	result, err := s.repo.ModerateIssueStatus(ctx, id, "rejected", -2)
	if err != nil {
		return err
	}
	if result.StatusChanged {
		s.publishStatusUpdated(ctx, id, result.PreviousStatus, "rejected")
	}
	s.recordModerationAction(ctx, "reject_issue", "issue", id, username, reason)
	return nil
}

func (s *adminService) BanDevice(ctx context.Context, id uuid.UUID, username string, reason *string) error {
	if err := s.repo.BanDevice(ctx, id, reason); err != nil {
		return err
	}
	s.recordModerationAction(ctx, "ban_device", "device", id, username, reason)
	return nil
}

func (s *adminService) publishStatusUpdated(ctx context.Context, id uuid.UUID, previousStatus, status string) {
	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), moderationAuditTimeout)
	defer cancel()

	if err := s.repo.PublishIssueStatusUpdated(auditCtx, id, previousStatus, status); err != nil {
		log.Printf("[ADMIN] status_event_failed issue=%s from=%s to=%s err=%v", id, previousStatus, status, err)
	}
}

func (s *adminService) recordModerationAction(
	ctx context.Context,
	actionType string,
	targetType string,
	targetID uuid.UUID,
	adminUsername string,
	note *string,
) {
	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), moderationAuditTimeout)
	defer cancel()

	if err := s.repo.CreateModerationAction(auditCtx, actionType, targetType, targetID, adminUsername, note); err != nil {
		log.Printf("[ADMIN] moderation_log_failed action=%s target_type=%s target_id=%s err=%v", actionType, targetType, targetID, err)
	}
}

func constantTimeSecretMatch(input, expected string) bool {
	inputHash := sha256.Sum256([]byte(input))
	expectedHash := sha256.Sum256([]byte(expected))
	return subtle.ConstantTimeCompare(inputHash[:], expectedHash[:]) == 1
}
