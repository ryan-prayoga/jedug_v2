package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/repository"
)

var ErrDeviceBanned = errors.New("device is banned")

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
	AnonToken     string
	Longitude     float64
	Latitude      float64
	GPSAccuracyM  *float64
	CapturedAt    *time.Time
	Severity      int
	HasCasualty   bool
	CasualtyCount int
	Note          *string
	Media         []MediaInput
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
		return nil, ErrDeviceBanned
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
		DeviceID:      device.ID,
		Longitude:     req.Longitude,
		Latitude:      req.Latitude,
		GPSAccuracyM:  req.GPSAccuracyM,
		CapturedAt:    req.CapturedAt,
		Severity:      req.Severity,
		HasCasualty:   req.HasCasualty,
		CasualtyCount: req.CasualtyCount,
		Note:          req.Note,
		Media:         mediaInputs,
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


