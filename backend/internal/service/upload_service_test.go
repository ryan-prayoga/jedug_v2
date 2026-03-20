package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
	"jedug_backend/internal/storage"
)

type uploadTestDeviceRepo struct {
	device *domain.Device
}

func (r *uploadTestDeviceRepo) FindByTokenHash(_ context.Context, tokenHash string) (*domain.Device, error) {
	if r.device == nil || r.device.AnonTokenHash != tokenHash {
		return nil, nil
	}
	return r.device, nil
}

func (*uploadTestDeviceRepo) Create(context.Context, *domain.Device) error { return nil }

func (*uploadTestDeviceRepo) UpdateLastSeen(context.Context, uuid.UUID) error { return nil }

func (*uploadTestDeviceRepo) CreateConsent(context.Context, *domain.DeviceConsent) error { return nil }

func (*uploadTestDeviceRepo) UpdateTrustScore(context.Context, uuid.UUID, int) error { return nil }

func (*uploadTestDeviceRepo) FindLastSubmissionTime(context.Context, uuid.UUID) (*time.Time, error) {
	return nil, nil
}

type uploadTestReportRepo struct {
	used map[string]bool
}

func (*uploadTestReportRepo) SubmitReport(context.Context, repository.SubmitInput) (*repository.SubmitResult, error) {
	return nil, errors.New("not implemented")
}

func (*uploadTestReportRepo) FindByClientRequestID(context.Context, uuid.UUID, uuid.UUID) (*repository.SubmitResult, error) {
	return nil, nil
}

func (r *uploadTestReportRepo) HasSubmissionMediaObjectKey(_ context.Context, objectKey string) (bool, error) {
	return r.used[objectKey], nil
}

type uploadTestTicketRepo struct {
	ticketsByKey       map[string]*repository.ReportUploadTicket
	pendingCount       int
	lastCreated        *repository.CreateReportUploadTicketInput
	countPendingCalled bool
}

func (r *uploadTestTicketRepo) CreateOrReplace(_ context.Context, input repository.CreateReportUploadTicketInput) error {
	copied := input
	r.lastCreated = &copied
	if r.ticketsByKey == nil {
		r.ticketsByKey = make(map[string]*repository.ReportUploadTicket)
	}
	r.ticketsByKey[input.ObjectKey] = &repository.ReportUploadTicket{
		ObjectKey:   input.ObjectKey,
		DeviceID:    input.DeviceID,
		ContentType: input.ContentType,
		SizeBytes:   input.SizeBytes,
		UploadMode:  input.UploadMode,
		IssuedAt:    time.Now().UTC(),
		ExpiresAt:   input.ExpiresAt,
	}
	return nil
}

func (r *uploadTestTicketRepo) FindByObjectKey(_ context.Context, objectKey string) (*repository.ReportUploadTicket, error) {
	if r.ticketsByKey == nil {
		return nil, nil
	}
	return r.ticketsByKey[objectKey], nil
}

func (r *uploadTestTicketRepo) CountPendingByDeviceSince(_ context.Context, _ uuid.UUID, _ time.Time) (int, error) {
	r.countPendingCalled = true
	return r.pendingCount, nil
}

func (*uploadTestTicketRepo) ListOrphansBefore(context.Context, time.Time, int) ([]repository.OrphanedReportUpload, error) {
	return nil, nil
}

func (*uploadTestTicketRepo) DeleteByObjectKeys(context.Context, []string) (int64, error) {
	return 0, nil
}

func (*uploadTestTicketRepo) CountOrphansBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type uploadTestDriver struct {
	name      string
	statByKey map[string]*storage.ObjectInfo
}

func (d *uploadTestDriver) Name() string { return d.name }

func (*uploadTestDriver) GenerateObjectKey(_ storage.UploadRequest) (string, error) {
	return "issues/2026/03/test.webp", nil
}

func (d *uploadTestDriver) CreatePresign(_ context.Context, req storage.UploadRequest, objectKey string) (*storage.PresignResult, error) {
	return &storage.PresignResult{
		ObjectKey:    objectKey,
		UploadMode:   d.name,
		UploadURL:    "/api/v1/uploads/file/" + objectKey,
		UploadMethod: "POST",
		PublicURL:    "https://cdn.example.test/" + objectKey,
		Headers: map[string]string{
			"Content-Type": storage.NormalizeContentType(req.ContentType),
		},
	}, nil
}

func (*uploadTestDriver) BuildPublicURL(objectKey string) string {
	return "https://cdn.example.test/" + objectKey
}

func (*uploadTestDriver) Upload(context.Context, string, string, []byte) error { return nil }

func (d *uploadTestDriver) Stat(_ context.Context, objectKey string) (*storage.ObjectInfo, error) {
	return d.statByKey[objectKey], nil
}

func (*uploadTestDriver) Delete(context.Context, string) error { return nil }

func TestUploadServiceCreateReportUploadRequiresKnownDevice(t *testing.T) {
	t.Parallel()

	store := storage.NewService(&uploadTestDriver{name: "local", statByKey: map[string]*storage.ObjectInfo{}}, nil)
	svc := NewUploadService(
		&uploadTestDeviceRepo{},
		&uploadTestReportRepo{used: map[string]bool{}},
		&uploadTestTicketRepo{},
		store,
		UploadServiceConfig{Secret: []byte("12345678901234567890123456789012")},
	)

	_, err := svc.CreateReportUpload(context.Background(), CreateReportUploadRequest{
		AnonToken:   "unknown",
		Filename:    "photo.webp",
		ContentType: "image/webp",
		SizeBytes:   1024,
	})
	if !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("expected ErrDeviceNotFound, got %v", err)
	}
}

func TestUploadServiceValidateLocalUploadChecksTicketBinding(t *testing.T) {
	t.Parallel()

	device := &domain.Device{
		ID:            uuid.New(),
		AnonTokenHash: hashToken("anon-token"),
	}
	driver := &uploadTestDriver{name: "local", statByKey: map[string]*storage.ObjectInfo{}}
	store := storage.NewService(driver, nil)
	ticketRepo := &uploadTestTicketRepo{}
	svc := NewUploadService(
		&uploadTestDeviceRepo{device: device},
		&uploadTestReportRepo{used: map[string]bool{}},
		ticketRepo,
		store,
		UploadServiceConfig{Secret: []byte("12345678901234567890123456789012")},
	)

	result, err := svc.CreateReportUpload(context.Background(), CreateReportUploadRequest{
		AnonToken:   "anon-token",
		Filename:    "photo.webp",
		ContentType: "image/webp",
		SizeBytes:   2048,
	})
	if err != nil {
		t.Fatalf("CreateReportUpload returned error: %v", err)
	}

	if err := svc.ValidateLocalUpload(context.Background(), result.UploadToken, result.ObjectKey, "image/webp", 2048); err != nil {
		t.Fatalf("expected valid local upload ticket, got %v", err)
	}
	if !ticketRepo.countPendingCalled {
		t.Fatalf("expected pending upload count to be checked during presign")
	}

	if err := svc.ValidateLocalUpload(context.Background(), result.UploadToken, result.ObjectKey, "image/png", 2048); !errors.Is(err, ErrUploadTokenInvalid) {
		t.Fatalf("expected ErrUploadTokenInvalid for mime mismatch, got %v", err)
	}
}

func TestUploadServiceValidateReportMediaRejectsMissingOrReusedObject(t *testing.T) {
	t.Parallel()

	device := &domain.Device{
		ID:            uuid.New(),
		AnonTokenHash: hashToken("anon-token"),
	}
	objectKey := "issues/2026/03/test.webp"
	driver := &uploadTestDriver{
		name: "local",
		statByKey: map[string]*storage.ObjectInfo{
			objectKey: &storage.ObjectInfo{
				SizeBytes: 2048,
			},
		},
	}
	reportRepo := &uploadTestReportRepo{used: map[string]bool{}}
	store := storage.NewService(driver, nil)
	ticketRepo := &uploadTestTicketRepo{}
	svc := NewUploadService(
		&uploadTestDeviceRepo{device: device},
		reportRepo,
		ticketRepo,
		store,
		UploadServiceConfig{Secret: []byte("12345678901234567890123456789012")},
	)

	result, err := svc.CreateReportUpload(context.Background(), CreateReportUploadRequest{
		AnonToken:   "anon-token",
		Filename:    "photo.webp",
		ContentType: "image/webp",
		SizeBytes:   2048,
	})
	if err != nil {
		t.Fatalf("CreateReportUpload returned error: %v", err)
	}

	err = svc.ValidateReportMedia(context.Background(), device.ID, []ReportMediaProof{{
		ObjectKey:   result.ObjectKey,
		MimeType:    "image/webp",
		SizeBytes:   2048,
		UploadToken: result.UploadToken,
	}})
	if err != nil {
		t.Fatalf("expected valid report media, got %v", err)
	}

	delete(driver.statByKey, objectKey)
	err = svc.ValidateReportMedia(context.Background(), device.ID, []ReportMediaProof{{
		ObjectKey:   result.ObjectKey,
		MimeType:    "image/webp",
		SizeBytes:   2048,
		UploadToken: result.UploadToken,
	}})
	if !errors.Is(err, ErrUploadObjectMissing) {
		t.Fatalf("expected ErrUploadObjectMissing, got %v", err)
	}

	driver.statByKey[objectKey] = &storage.ObjectInfo{SizeBytes: 2048}
	reportRepo.used[objectKey] = true
	err = svc.ValidateReportMedia(context.Background(), device.ID, []ReportMediaProof{{
		ObjectKey:   result.ObjectKey,
		MimeType:    "image/webp",
		SizeBytes:   2048,
		UploadToken: result.UploadToken,
	}})
	if !errors.Is(err, ErrUploadAlreadyUsed) {
		t.Fatalf("expected ErrUploadAlreadyUsed, got %v", err)
	}
}

func TestUploadServiceCreateReportUploadRejectsTooManyPendingTickets(t *testing.T) {
	t.Parallel()

	device := &domain.Device{
		ID:            uuid.New(),
		AnonTokenHash: hashToken("anon-token"),
	}
	store := storage.NewService(&uploadTestDriver{name: "local", statByKey: map[string]*storage.ObjectInfo{}}, nil)
	svc := NewUploadService(
		&uploadTestDeviceRepo{device: device},
		&uploadTestReportRepo{used: map[string]bool{}},
		&uploadTestTicketRepo{pendingCount: 4},
		store,
		UploadServiceConfig{
			Secret:       []byte("12345678901234567890123456789012"),
			PendingLimit: 4,
		},
	)

	_, err := svc.CreateReportUpload(context.Background(), CreateReportUploadRequest{
		AnonToken:   "anon-token",
		Filename:    "photo.webp",
		ContentType: "image/webp",
		SizeBytes:   1024,
	})
	if !errors.Is(err, ErrUploadPendingLimitReached) {
		t.Fatalf("expected ErrUploadPendingLimitReached, got %v", err)
	}
}
