package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/repository"
)

var (
	ErrDeviceBanned  = errors.New("device is banned")
	ErrCooldown      = errors.New("submission cooldown active")
	ErrLowTrustScore = errors.New("device trust score too low")
)

const (
	submitCooldown = 2 * time.Minute
	minTrustScore  = -10
)

type ReportService interface {
	SubmitReport(ctx context.Context, req SubmitReportRequest) (*SubmitReportResult, error)
}

type MediaInput struct {
	ObjectKey string
	MimeType  string
	SizeBytes int
	Width     *int
	Height    *int
	SHA256    *string
	IsPrimary bool
	SortOrder int
}

type SubmitReportRequest struct {
	ClientRequestID *string
	AnonToken       string
	Longitude       float64
	Latitude        float64
	GPSAccuracyM    *float64
	CapturedAt      *time.Time
	Severity        int
	HasCasualty     bool
	CasualtyCount   int
	Note            *string
	Media           []MediaInput
}

type SubmitReportResult struct {
	IssueID      uuid.UUID `json:"issue_id"`
	SubmissionID uuid.UUID `json:"submission_id"`
	IsNewIssue   bool      `json:"is_new_issue"`
}

type reportService struct {
	deviceRepo repository.DeviceRepository
	reportRepo repository.ReportRepository
}

func NewReportService(deviceRepo repository.DeviceRepository, reportRepo repository.ReportRepository) ReportService {
	return &reportService{
		deviceRepo: deviceRepo,
		reportRepo: reportRepo,
	}
}

func (s *reportService) SubmitReport(ctx context.Context, req SubmitReportRequest) (*SubmitReportResult, error) {
	tokenHash := hashToken(req.AnonToken)
	device, err := s.deviceRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, ErrDeviceNotFound
	}
	if device.IsBanned {
		log.Printf("[ANTISPAM] banned_submit device=%s", device.ID)
		return nil, ErrDeviceBanned
	}

	// Trust score check
	if device.TrustScore <= minTrustScore {
		log.Printf("[ANTISPAM] low_trust device=%s score=%d", device.ID, device.TrustScore)
		return nil, ErrLowTrustScore
	}

	// Cooldown check
	lastSubmit, err := s.deviceRepo.FindLastSubmissionTime(ctx, device.ID)
	if err != nil {
		return nil, err
	}
	if lastSubmit != nil && time.Since(*lastSubmit) < submitCooldown {
		log.Printf("[ANTISPAM] cooldown device=%s last_submit=%s", device.ID, lastSubmit.Format(time.RFC3339))
		return nil, ErrCooldown
	}

	// Idempotency: resolve or generate client_request_id
	var clientRequestID uuid.UUID
	if req.ClientRequestID != nil && *req.ClientRequestID != "" {
		parsed, parseErr := uuid.Parse(*req.ClientRequestID)
		if parseErr != nil {
			return nil, errors.New("invalid client_request_id format")
		}
		clientRequestID = parsed

		existing, findErr := s.reportRepo.FindByClientRequestID(ctx, clientRequestID)
		if findErr != nil {
			return nil, findErr
		}
		if existing != nil {
			log.Printf("[ANTISPAM] idempotent_return device=%s client_request_id=%s", device.ID, clientRequestID)
			return &SubmitReportResult{
				IssueID:      existing.IssueID,
				SubmissionID: existing.SubmissionID,
				IsNewIssue:   existing.IsNewIssue,
			}, nil
		}
	} else {
		clientRequestID = uuid.New()
	}

	mediaInputs := make([]repository.SubmitMediaInput, len(req.Media))
	for i, m := range req.Media {
		mediaInputs[i] = repository.SubmitMediaInput{
			ObjectKey: m.ObjectKey,
			MimeType:  m.MimeType,
			SizeBytes: m.SizeBytes,
			Width:     m.Width,
			Height:    m.Height,
			SHA256:    m.SHA256,
			IsPrimary: m.IsPrimary,
			SortOrder: m.SortOrder,
		}
	}

	result, err := s.reportRepo.SubmitReport(ctx, repository.SubmitInput{
		ClientRequestID: clientRequestID,
		DeviceID:        device.ID,
		Longitude:       req.Longitude,
		Latitude:        req.Latitude,
		GPSAccuracyM:    req.GPSAccuracyM,
		CapturedAt:      req.CapturedAt,
		Severity:        req.Severity,
		HasCasualty:     req.HasCasualty,
		CasualtyCount:   req.CasualtyCount,
		Note:            req.Note,
		Media:           mediaInputs,
	})
	if err != nil {
		return nil, err
	}

	return &SubmitReportResult{
		IssueID:      result.IssueID,
		SubmissionID: result.SubmissionID,
		IsNewIssue:   result.IsNewIssue,
	}, nil
}


