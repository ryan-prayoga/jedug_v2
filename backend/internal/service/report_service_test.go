package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

type reportTestDeviceRepo struct {
	device       *domain.Device
	lastSubmit   *time.Time
	lastSubmitID uuid.UUID
}

func (r *reportTestDeviceRepo) FindByTokenHash(_ context.Context, tokenHash string) (*domain.Device, error) {
	if r.device == nil || r.device.AnonTokenHash != tokenHash {
		return nil, nil
	}
	return r.device, nil
}

func (*reportTestDeviceRepo) Create(context.Context, *domain.Device) error { return nil }

func (*reportTestDeviceRepo) UpdateLastSeen(context.Context, uuid.UUID) error { return nil }

func (*reportTestDeviceRepo) CreateConsent(context.Context, *domain.DeviceConsent) error { return nil }

func (*reportTestDeviceRepo) UpdateTrustScore(context.Context, uuid.UUID, int) error { return nil }

func (r *reportTestDeviceRepo) FindLastSubmissionTime(_ context.Context, deviceID uuid.UUID) (*time.Time, error) {
	r.lastSubmitID = deviceID
	return r.lastSubmit, nil
}

type reportTestReportRepo struct {
	existing     *repository.SubmitResult
	findCalls    int
	submitCalls  int
	lastFindArgs struct {
		deviceID        uuid.UUID
		clientRequestID uuid.UUID
	}
	lastSubmitInput repository.SubmitInput
	submitResult    *repository.SubmitResult
	submitErr       error
}

func (r *reportTestReportRepo) SubmitReport(_ context.Context, input repository.SubmitInput) (*repository.SubmitResult, error) {
	r.submitCalls++
	r.lastSubmitInput = input
	if r.submitErr != nil {
		return nil, r.submitErr
	}
	if r.submitResult != nil {
		return r.submitResult, nil
	}
	return &repository.SubmitResult{
		IssueID:      uuid.New(),
		SubmissionID: uuid.New(),
		IsNewIssue:   false,
	}, nil
}

func (r *reportTestReportRepo) FindByClientRequestID(_ context.Context, deviceID, clientRequestID uuid.UUID) (*repository.SubmitResult, error) {
	r.findCalls++
	r.lastFindArgs.deviceID = deviceID
	r.lastFindArgs.clientRequestID = clientRequestID
	return r.existing, nil
}

func (*reportTestReportRepo) HasSubmissionMediaObjectKey(context.Context, string) (bool, error) {
	return false, nil
}

type reportTestUploadSvc struct {
	validateCalls int
}

func (*reportTestUploadSvc) CreateReportUpload(context.Context, CreateReportUploadRequest) (*CreateReportUploadResult, error) {
	return nil, errors.New("not implemented")
}

func (*reportTestUploadSvc) ValidateLocalUpload(context.Context, string, string, string, int) error {
	return errors.New("not implemented")
}

func (s *reportTestUploadSvc) ValidateReportMedia(context.Context, uuid.UUID, []ReportMediaProof) error {
	s.validateCalls++
	return nil
}

type reportTestLocationNormalizer struct{}

func (*reportTestLocationNormalizer) NormalizeForReport(context.Context, float64, float64) ReportLocationNormalization {
	return ReportLocationNormalization{}
}

func TestReportServiceReturnsExistingSubmissionBeforeCooldown(t *testing.T) {
	t.Parallel()

	clientRequestID := uuid.New()
	requestIDRaw := clientRequestID.String()
	deviceID := uuid.New()
	lastSubmit := time.Now().Add(-30 * time.Second)
	req := SubmitReportRequest{
		ClientRequestID: &requestIDRaw,
		AnonToken:       "anon-token",
		Latitude:        -6.2,
		Longitude:       106.8,
		Severity:        3,
		HasCasualty:     false,
		CasualtyCount:   0,
		Media: []MediaInput{{
			ObjectKey:   "issues/2026/03/a.webp",
			MimeType:    "image/webp",
			SizeBytes:   2048,
			UploadToken: "ticket",
			IsPrimary:   true,
			SortOrder:   0,
		}},
	}
	fingerprint, err := buildReportRequestFingerprint(req)
	if err != nil {
		t.Fatalf("buildReportRequestFingerprint returned error: %v", err)
	}

	reportRepo := &reportTestReportRepo{
		existing: &repository.SubmitResult{
			IssueID:            uuid.New(),
			SubmissionID:       uuid.New(),
			IsNewIssue:         true,
			RequestFingerprint: fingerprint,
		},
	}
	uploadSvc := &reportTestUploadSvc{}
	svc := NewReportService(
		&reportTestDeviceRepo{
			device: &domain.Device{
				ID:            deviceID,
				AnonTokenHash: hashToken("anon-token"),
			},
			lastSubmit: &lastSubmit,
		},
		reportRepo,
		&reportTestLocationNormalizer{},
		uploadSvc,
	)

	result, err := svc.SubmitReport(context.Background(), req)
	if err != nil {
		t.Fatalf("SubmitReport returned error: %v", err)
	}
	if result == nil {
		t.Fatalf("SubmitReport returned nil result")
	}
	if result.IssueID != reportRepo.existing.IssueID || result.SubmissionID != reportRepo.existing.SubmissionID {
		t.Fatalf("SubmitReport returned unexpected existing submission result")
	}
	if !result.IsNewIssue {
		t.Fatalf("expected IsNewIssue to match stored submission")
	}
	if reportRepo.submitCalls != 0 {
		t.Fatalf("expected no new repository submit, got %d", reportRepo.submitCalls)
	}
	if uploadSvc.validateCalls != 0 {
		t.Fatalf("expected no upload validation on idempotent replay, got %d", uploadSvc.validateCalls)
	}
}

func TestReportServiceRejectsIdempotencyKeyPayloadMismatch(t *testing.T) {
	t.Parallel()

	clientRequestID := uuid.New()
	requestIDRaw := clientRequestID.String()
	reportRepo := &reportTestReportRepo{
		existing: &repository.SubmitResult{
			IssueID:            uuid.New(),
			SubmissionID:       uuid.New(),
			IsNewIssue:         false,
			RequestFingerprint: "different-fingerprint",
		},
	}
	svc := NewReportService(
		&reportTestDeviceRepo{
			device: &domain.Device{
				ID:            uuid.New(),
				AnonTokenHash: hashToken("anon-token"),
			},
		},
		reportRepo,
		&reportTestLocationNormalizer{},
		&reportTestUploadSvc{},
	)

	_, err := svc.SubmitReport(context.Background(), SubmitReportRequest{
		ClientRequestID: &requestIDRaw,
		AnonToken:       "anon-token",
		Latitude:        -6.2,
		Longitude:       106.8,
		Severity:        3,
		HasCasualty:     false,
		CasualtyCount:   0,
		Media: []MediaInput{{
			ObjectKey:   "issues/2026/03/a.webp",
			MimeType:    "image/webp",
			SizeBytes:   2048,
			UploadToken: "ticket",
			IsPrimary:   true,
			SortOrder:   0,
		}},
	})
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected ErrIdempotencyConflict, got %v", err)
	}
	if reportRepo.submitCalls != 0 {
		t.Fatalf("expected no new repository submit, got %d", reportRepo.submitCalls)
	}
}

func TestReportServiceMapsRepositoryIdempotencyConflict(t *testing.T) {
	t.Parallel()

	clientRequestID := uuid.New()
	requestIDRaw := clientRequestID.String()
	uploadSvc := &reportTestUploadSvc{}
	reportRepo := &reportTestReportRepo{
		submitErr: repository.ErrReportIdempotencyConflict,
	}
	svc := NewReportService(
		&reportTestDeviceRepo{
			device: &domain.Device{
				ID:            uuid.New(),
				AnonTokenHash: hashToken("anon-token"),
			},
		},
		reportRepo,
		&reportTestLocationNormalizer{},
		uploadSvc,
	)

	_, err := svc.SubmitReport(context.Background(), SubmitReportRequest{
		ClientRequestID: &requestIDRaw,
		AnonToken:       "anon-token",
		Latitude:        -6.2,
		Longitude:       106.8,
		Severity:        3,
		HasCasualty:     false,
		CasualtyCount:   0,
		Media: []MediaInput{{
			ObjectKey:   "issues/2026/03/a.webp",
			MimeType:    "image/webp",
			SizeBytes:   2048,
			UploadToken: "ticket",
			IsPrimary:   true,
			SortOrder:   0,
		}},
	})
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected ErrIdempotencyConflict, got %v", err)
	}
	if reportRepo.submitCalls != 1 {
		t.Fatalf("expected repository submit to be attempted once, got %d", reportRepo.submitCalls)
	}
	if uploadSvc.validateCalls != 1 {
		t.Fatalf("expected upload validation before repository submit, got %d", uploadSvc.validateCalls)
	}
}
