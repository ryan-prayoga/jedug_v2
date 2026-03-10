package service

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/google/uuid"
	"jedug_backend/internal/repository"
)

var (
	ErrAlreadyFlagged    = errors.New("already flagged")
	ErrInvalidFlagReason = errors.New("invalid flag reason")
)

const autoHideFlagThreshold = 3

var validFlagReasons = map[string]bool{
	"spam": true, "hoax": true, "off_topic": true,
	"duplicate": true, "abuse": true, "other": true,
}

type FlagIssueRequest struct {
	AnonToken string
	IssueID   uuid.UUID
	Reason    string
	Note      *string
}

type FlagService interface {
	FlagIssue(ctx context.Context, req FlagIssueRequest) error
}

type flagService struct {
	deviceRepo repository.DeviceRepository
	flagRepo   repository.FlagRepository
	adminRepo  repository.AdminRepository
}

func NewFlagService(deviceRepo repository.DeviceRepository, flagRepo repository.FlagRepository, adminRepo repository.AdminRepository) FlagService {
	return &flagService{
		deviceRepo: deviceRepo,
		flagRepo:   flagRepo,
		adminRepo:  adminRepo,
	}
}

func (s *flagService) FlagIssue(ctx context.Context, req FlagIssueRequest) error {
	if !validFlagReasons[req.Reason] {
		return ErrInvalidFlagReason
	}

	tokenHash := hashToken(req.AnonToken)
	device, err := s.deviceRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}
	if device == nil {
		return ErrDeviceNotFound
	}
	if device.IsBanned {
		return ErrDeviceBanned
	}

	err = s.flagRepo.CreateIssueFlag(ctx, req.IssueID, device.ID, req.Reason, req.Note)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateFlag) {
			return ErrAlreadyFlagged
		}
		return err
	}

	// Check unique flag count for auto-hide
	uniqueCount, err := s.flagRepo.CountUniqueIssueFlags(ctx, req.IssueID)
	if err != nil {
		log.Printf("[ANTISPAM] failed to count flags issue=%s err=%v", req.IssueID, err)
		return nil
	}

	if uniqueCount >= autoHideFlagThreshold {
		reason := "auto-hidden: " + strconv.Itoa(uniqueCount) + " unique device flags"
		if err := s.adminRepo.UpdateIssueHidden(ctx, req.IssueID, true, &reason); err != nil {
			log.Printf("[ANTISPAM] auto_hide_failed issue=%s err=%v", req.IssueID, err)
			return nil
		}
		if err := s.adminRepo.CreateModerationAction(ctx, "auto_hide_issue", "issue", req.IssueID, "system", &reason); err != nil {
			log.Printf("[ANTISPAM] auto_hide_log_failed issue=%s err=%v", req.IssueID, err)
		}
		log.Printf("[ANTISPAM] auto_hide issue=%s unique_flags=%d", req.IssueID, uniqueCount)
	}

	return nil
}
