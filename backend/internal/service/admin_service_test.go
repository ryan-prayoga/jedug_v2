package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

type adminRepoFake struct {
	updateIssueHiddenErr error
	moderateIssueErr     error
	moderateIssueResult  *repository.IssueStatusModerationResult
	publishStatusErr     error
	banDeviceErr         error
	createActionErr      error

	updateIssueHiddenCalls int
	moderateIssueCalls     int
	publishStatusCalls     int
	banDeviceCalls         int
	createActionCalls      int

	lastModeratedIssueID uuid.UUID
	lastModeratedStatus  string
	lastTrustDelta       int
	lastPublishedIssueID uuid.UUID
	lastPublishedFrom    string
	lastPublishedTo      string
	lastActionType       string
	lastActionTargetType string
	lastActionTargetID   uuid.UUID
}

func (f *adminRepoFake) ListIssues(_ context.Context, _ int, _ int, _ *string) ([]*domain.AdminIssue, error) {
	return nil, nil
}

func (f *adminRepoFake) FindIssueByID(_ context.Context, _ uuid.UUID) (*domain.AdminIssue, error) {
	return nil, nil
}

func (f *adminRepoFake) FindIssueByIDWithDetail(_ context.Context, _ uuid.UUID) (*domain.AdminIssueDetail, error) {
	return nil, nil
}

func (f *adminRepoFake) UpdateIssueHidden(_ context.Context, _ uuid.UUID, _ bool, _ *string) error {
	f.updateIssueHiddenCalls++
	return f.updateIssueHiddenErr
}

func (f *adminRepoFake) ModerateIssueStatus(_ context.Context, id uuid.UUID, status string, trustDelta int) (*repository.IssueStatusModerationResult, error) {
	f.moderateIssueCalls++
	f.lastModeratedIssueID = id
	f.lastModeratedStatus = status
	f.lastTrustDelta = trustDelta
	if f.moderateIssueErr != nil {
		return nil, f.moderateIssueErr
	}
	if f.moderateIssueResult == nil {
		return &repository.IssueStatusModerationResult{}, nil
	}
	return f.moderateIssueResult, nil
}

func (f *adminRepoFake) PublishIssueStatusUpdated(_ context.Context, id uuid.UUID, previousStatus, status string) error {
	f.publishStatusCalls++
	f.lastPublishedIssueID = id
	f.lastPublishedFrom = previousStatus
	f.lastPublishedTo = status
	return f.publishStatusErr
}

func (f *adminRepoFake) BanDevice(_ context.Context, _ uuid.UUID, _ *string) error {
	f.banDeviceCalls++
	return f.banDeviceErr
}

func (f *adminRepoFake) CreateModerationAction(_ context.Context, actionType, targetType string, targetID uuid.UUID, _ string, _ *string) error {
	f.createActionCalls++
	f.lastActionType = actionType
	f.lastActionTargetType = targetType
	f.lastActionTargetID = targetID
	return f.createActionErr
}

func (f *adminRepoFake) GetModerationLog(_ context.Context, _ string, _ uuid.UUID) ([]*domain.ModerationAction, error) {
	return nil, nil
}

func (f *adminRepoFake) AdjustSubmitterTrustScores(_ context.Context, _ uuid.UUID, _ int) error {
	return nil
}

func TestAdminServiceFixIssueDoesNotFailOnAuditErrors(t *testing.T) {
	repoFake := &adminRepoFake{
		moderateIssueResult: &repository.IssueStatusModerationResult{
			PreviousStatus: "open",
			StatusChanged:  true,
		},
		publishStatusErr: errors.New("issue events unavailable"),
		createActionErr:  errors.New("moderation actions unavailable"),
	}
	svc := NewAdminService("admin", "secret", repoFake)
	issueID := uuid.New()

	err := svc.FixIssue(context.Background(), issueID, "admin", nil)
	if err != nil {
		t.Fatalf("FixIssue returned error: %v", err)
	}
	if repoFake.moderateIssueCalls != 1 {
		t.Fatalf("expected ModerateIssueStatus to be called once, got %d", repoFake.moderateIssueCalls)
	}
	if repoFake.lastModeratedStatus != "fixed" {
		t.Fatalf("unexpected status: %s", repoFake.lastModeratedStatus)
	}
	if repoFake.lastTrustDelta != 1 {
		t.Fatalf("unexpected trust delta: %d", repoFake.lastTrustDelta)
	}
	if repoFake.publishStatusCalls != 1 {
		t.Fatalf("expected PublishIssueStatusUpdated to be called once, got %d", repoFake.publishStatusCalls)
	}
	if repoFake.lastPublishedIssueID != issueID || repoFake.lastPublishedFrom != "open" || repoFake.lastPublishedTo != "fixed" {
		t.Fatalf("unexpected published status event: issue=%s from=%s to=%s", repoFake.lastPublishedIssueID, repoFake.lastPublishedFrom, repoFake.lastPublishedTo)
	}
	if repoFake.createActionCalls != 1 {
		t.Fatalf("expected CreateModerationAction to be called once, got %d", repoFake.createActionCalls)
	}
	if repoFake.lastActionType != "mark_fixed" || repoFake.lastActionTargetType != "issue" || repoFake.lastActionTargetID != issueID {
		t.Fatalf("unexpected moderation action payload: type=%s target_type=%s target_id=%s", repoFake.lastActionType, repoFake.lastActionTargetType, repoFake.lastActionTargetID)
	}
}

func TestAdminServiceRejectIssueReturnsNotFound(t *testing.T) {
	repoFake := &adminRepoFake{
		moderateIssueErr: repository.ErrModerationTargetNotFound,
	}
	svc := NewAdminService("admin", "secret", repoFake)

	err := svc.RejectIssue(context.Background(), uuid.New(), "admin", nil)
	if !errors.Is(err, ErrModerationTargetNotFound) {
		t.Fatalf("expected ErrModerationTargetNotFound, got %v", err)
	}
	if repoFake.publishStatusCalls != 0 {
		t.Fatalf("expected PublishIssueStatusUpdated to be skipped, got %d calls", repoFake.publishStatusCalls)
	}
	if repoFake.createActionCalls != 0 {
		t.Fatalf("expected CreateModerationAction to be skipped, got %d calls", repoFake.createActionCalls)
	}
}

func TestAdminServiceRejectIssueSkipsStatusEventWhenAlreadyRejected(t *testing.T) {
	repoFake := &adminRepoFake{
		moderateIssueResult: &repository.IssueStatusModerationResult{
			PreviousStatus: "rejected",
			StatusChanged:  false,
		},
	}
	svc := NewAdminService("admin", "secret", repoFake)
	issueID := uuid.New()

	err := svc.RejectIssue(context.Background(), issueID, "admin", nil)
	if err != nil {
		t.Fatalf("RejectIssue returned error: %v", err)
	}
	if repoFake.publishStatusCalls != 0 {
		t.Fatalf("expected PublishIssueStatusUpdated to be skipped, got %d calls", repoFake.publishStatusCalls)
	}
	if repoFake.createActionCalls != 1 {
		t.Fatalf("expected CreateModerationAction to be called once, got %d", repoFake.createActionCalls)
	}
}

func TestAdminServiceBanDeviceDoesNotFailOnAuditError(t *testing.T) {
	repoFake := &adminRepoFake{
		createActionErr: errors.New("moderation actions unavailable"),
	}
	svc := NewAdminService("admin", "secret", repoFake)

	err := svc.BanDevice(context.Background(), uuid.New(), "admin", nil)
	if err != nil {
		t.Fatalf("BanDevice returned error: %v", err)
	}
	if repoFake.banDeviceCalls != 1 {
		t.Fatalf("expected BanDevice repository call once, got %d", repoFake.banDeviceCalls)
	}
	if repoFake.createActionCalls != 1 {
		t.Fatalf("expected CreateModerationAction to be called once, got %d", repoFake.createActionCalls)
	}
}
