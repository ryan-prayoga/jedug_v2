package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/repository"
	"jedug_backend/internal/storage"
)

var (
	ErrUploadTokenRequired         = errors.New("upload token is required")
	ErrUploadTokenInvalid          = errors.New("upload token is invalid")
	ErrUploadTokenExpired          = errors.New("upload token is expired")
	ErrUploadOwnershipMismatch     = errors.New("upload ownership mismatch")
	ErrUploadAlreadyUsed           = errors.New("upload already used")
	ErrUploadObjectMissing         = errors.New("uploaded object is missing")
	ErrUploadedObjectMismatch      = errors.New("uploaded object does not match ticket")
	ErrUploadDeviceBootstrapNeeded = errors.New("device bootstrap is required")
	ErrUploadPendingLimitReached   = errors.New("too many pending uploads for device")
)

const (
	defaultUploadTicketTTL     = 10 * time.Minute
	defaultUploadPendingWindow = 30 * time.Minute
	defaultUploadPendingLimit  = 4
)

type UploadService interface {
	CreateReportUpload(ctx context.Context, req CreateReportUploadRequest) (*CreateReportUploadResult, error)
	ValidateLocalUpload(ctx context.Context, uploadToken, objectKey, contentType string, sizeBytes int) error
	ValidateReportMedia(ctx context.Context, deviceID uuid.UUID, media []ReportMediaProof) error
}

type CreateReportUploadRequest struct {
	AnonToken   string
	Filename    string
	ContentType string
	SizeBytes   int
}

type CreateReportUploadResult struct {
	ObjectKey         string            `json:"object_key"`
	UploadMode        string            `json:"upload_mode"`
	UploadURL         string            `json:"upload_url"`
	UploadMethod      string            `json:"upload_method,omitempty"`
	PublicURL         string            `json:"public_url"`
	Headers           map[string]string `json:"headers,omitempty"`
	UploadToken       string            `json:"upload_token"`
	UploadExpiresAt   time.Time         `json:"upload_expires_at"`
	UploadTokenType   string            `json:"upload_token_type"`
	UploadTokenHeader string            `json:"upload_token_header,omitempty"`
}

type ReportMediaProof struct {
	ObjectKey   string
	MimeType    string
	SizeBytes   int
	UploadToken string
}

type UploadServiceConfig struct {
	Secret        []byte
	TTL           time.Duration
	PendingWindow time.Duration
	PendingLimit  int
}

type uploadService struct {
	deviceRepo       repository.DeviceRepository
	reportRepo       repository.ReportRepository
	uploadTicketRepo repository.ReportUploadTicketRepository
	storage          *storage.Service
	secret           []byte
	ttl              time.Duration
	pendingWindow    time.Duration
	pendingLimit     int
	now              func() time.Time
}

type uploadTicketClaims struct {
	Version     int    `json:"ver"`
	Purpose     string `json:"purpose"`
	DeviceID    string `json:"device_id"`
	ObjectKey   string `json:"object_key"`
	ContentType string `json:"content_type"`
	SizeBytes   int    `json:"size_bytes"`
	IssuedAt    int64  `json:"iat"`
	ExpiresAt   int64  `json:"exp"`
}

func NewUploadService(
	deviceRepo repository.DeviceRepository,
	reportRepo repository.ReportRepository,
	uploadTicketRepo repository.ReportUploadTicketRepository,
	store *storage.Service,
	cfg UploadServiceConfig,
) UploadService {
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = defaultUploadTicketTTL
	}
	pendingWindow := cfg.PendingWindow
	if pendingWindow <= 0 {
		pendingWindow = defaultUploadPendingWindow
	}
	pendingLimit := cfg.PendingLimit
	if pendingLimit <= 0 {
		pendingLimit = defaultUploadPendingLimit
	}

	return &uploadService{
		deviceRepo:       deviceRepo,
		reportRepo:       reportRepo,
		uploadTicketRepo: uploadTicketRepo,
		storage:          store,
		secret:           append([]byte(nil), cfg.Secret...),
		ttl:              ttl,
		pendingWindow:    pendingWindow,
		pendingLimit:     pendingLimit,
		now:              time.Now,
	}
}

func (s *uploadService) CreateReportUpload(ctx context.Context, req CreateReportUploadRequest) (*CreateReportUploadResult, error) {
	device, err := s.resolveDevice(ctx, req.AnonToken)
	if err != nil {
		return nil, err
	}
	if device.IsBanned {
		return nil, ErrDeviceBanned
	}
	if s.uploadTicketRepo != nil {
		pendingCount, err := s.uploadTicketRepo.CountPendingByDeviceSince(ctx, device.ID, s.now().Add(-s.pendingWindow))
		if err != nil {
			return nil, err
		}
		if pendingCount >= s.pendingLimit {
			return nil, ErrUploadPendingLimitReached
		}
	}

	presign, err := s.storage.CreatePresign(ctx, storage.UploadRequest{
		Filename:    req.Filename,
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
	})
	if err != nil {
		return nil, err
	}

	uploadToken, expiresAt, err := s.signTicket(device.ID, presign.ObjectKey, req.ContentType, req.SizeBytes)
	if err != nil {
		return nil, err
	}
	if s.uploadTicketRepo != nil {
		if err := s.uploadTicketRepo.CreateOrReplace(ctx, repository.CreateReportUploadTicketInput{
			ObjectKey:   presign.ObjectKey,
			DeviceID:    device.ID,
			ContentType: storage.NormalizeContentType(req.ContentType),
			SizeBytes:   req.SizeBytes,
			UploadMode:  presign.UploadMode,
			ExpiresAt:   expiresAt,
		}); err != nil {
			return nil, err
		}
	}

	return &CreateReportUploadResult{
		ObjectKey:         presign.ObjectKey,
		UploadMode:        presign.UploadMode,
		UploadURL:         presign.UploadURL,
		UploadMethod:      presign.UploadMethod,
		PublicURL:         presign.PublicURL,
		Headers:           presign.Headers,
		UploadToken:       uploadToken,
		UploadExpiresAt:   expiresAt,
		UploadTokenType:   "upload_ticket",
		UploadTokenHeader: "X-Upload-Token",
	}, nil
}

func (s *uploadService) ValidateLocalUpload(ctx context.Context, uploadToken, objectKey, contentType string, sizeBytes int) error {
	claims, err := s.authenticate(uploadToken)
	if err != nil {
		return err
	}
	if err := s.validateClaims(claims, uuid.Nil, objectKey, contentType, sizeBytes); err != nil {
		return err
	}
	if err := s.validateTicketRecord(ctx, claims, uuid.Nil, objectKey, contentType, sizeBytes); err != nil {
		return err
	}
	used, err := s.reportRepo.HasSubmissionMediaObjectKey(ctx, objectKey)
	if err != nil {
		return err
	}
	if used {
		return ErrUploadAlreadyUsed
	}
	return nil
}

func (s *uploadService) ValidateReportMedia(ctx context.Context, deviceID uuid.UUID, media []ReportMediaProof) error {
	seen := make(map[string]struct{}, len(media))
	for _, item := range media {
		if _, ok := seen[item.ObjectKey]; ok {
			return ErrUploadAlreadyUsed
		}
		seen[item.ObjectKey] = struct{}{}

		claims, err := s.authenticate(item.UploadToken)
		if err != nil {
			return err
		}
		if err := s.validateClaims(claims, deviceID, item.ObjectKey, item.MimeType, item.SizeBytes); err != nil {
			return err
		}
		if err := s.validateTicketRecord(ctx, claims, deviceID, item.ObjectKey, item.MimeType, item.SizeBytes); err != nil {
			return err
		}

		used, err := s.reportRepo.HasSubmissionMediaObjectKey(ctx, item.ObjectKey)
		if err != nil {
			return err
		}
		if used {
			return ErrUploadAlreadyUsed
		}

		info, err := s.storage.Stat(ctx, item.ObjectKey)
		if err != nil {
			return err
		}
		if info == nil {
			return ErrUploadObjectMissing
		}
		if info.SizeBytes != int64(item.SizeBytes) {
			return ErrUploadedObjectMismatch
		}
		if info.ContentType != "" && info.ContentType != storage.NormalizeContentType(item.MimeType) {
			return ErrUploadedObjectMismatch
		}
	}

	return nil
}

func (s *uploadService) resolveDevice(ctx context.Context, anonToken string) (*repositoryDevice, error) {
	anonToken = strings.TrimSpace(anonToken)
	if anonToken == "" {
		return nil, ErrUploadDeviceBootstrapNeeded
	}
	tokenHash := hashToken(anonToken)
	device, err := s.deviceRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, ErrDeviceNotFound
	}
	return &repositoryDevice{ID: device.ID, IsBanned: device.IsBanned}, nil
}

func (s *uploadService) signTicket(deviceID uuid.UUID, objectKey, contentType string, sizeBytes int) (string, time.Time, error) {
	now := s.now().UTC()
	expiresAt := now.Add(s.ttl)
	payload, err := json.Marshal(uploadTicketClaims{
		Version:     1,
		Purpose:     "report_media",
		DeviceID:    deviceID.String(),
		ObjectKey:   storage.NormalizeObjectKey(objectKey),
		ContentType: storage.NormalizeContentType(contentType),
		SizeBytes:   sizeBytes,
		IssuedAt:    now.Unix(),
		ExpiresAt:   expiresAt.Unix(),
	})
	if err != nil {
		return "", time.Time{}, err
	}

	payloadPart := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payloadPart))
	signaturePart := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return payloadPart + "." + signaturePart, expiresAt, nil
}

func (s *uploadService) authenticate(rawToken string) (*uploadTicketClaims, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, ErrUploadTokenRequired
	}

	parts := strings.Split(rawToken, ".")
	if len(parts) != 2 {
		return nil, ErrUploadTokenInvalid
	}

	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(parts[0]))
	expectedSig := mac.Sum(nil)

	actualSig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrUploadTokenInvalid
	}
	if subtle.ConstantTimeCompare(expectedSig, actualSig) != 1 {
		return nil, ErrUploadTokenInvalid
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrUploadTokenInvalid
	}

	var claims uploadTicketClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrUploadTokenInvalid
	}
	if claims.Version != 1 || claims.Purpose != "report_media" {
		return nil, ErrUploadTokenInvalid
	}
	if claims.ExpiresAt <= s.now().Unix() {
		return nil, ErrUploadTokenExpired
	}
	return &claims, nil
}

func (s *uploadService) validateClaims(
	claims *uploadTicketClaims,
	deviceID uuid.UUID,
	objectKey, contentType string,
	sizeBytes int,
) error {
	if claims == nil {
		return ErrUploadTokenInvalid
	}
	if deviceID != uuid.Nil && claims.DeviceID != deviceID.String() {
		return ErrUploadOwnershipMismatch
	}
	if storage.NormalizeObjectKey(objectKey) != claims.ObjectKey {
		return ErrUploadTokenInvalid
	}
	if storage.NormalizeContentType(contentType) != claims.ContentType {
		return ErrUploadTokenInvalid
	}
	if sizeBytes != claims.SizeBytes {
		return ErrUploadTokenInvalid
	}
	if err := storage.ValidateSubmittedMedia(objectKey, contentType, sizeBytes); err != nil {
		return fmt.Errorf("%w: %v", ErrUploadTokenInvalid, err)
	}
	return nil
}

func (s *uploadService) validateTicketRecord(
	ctx context.Context,
	claims *uploadTicketClaims,
	deviceID uuid.UUID,
	objectKey, contentType string,
	sizeBytes int,
) error {
	if s.uploadTicketRepo == nil {
		return nil
	}

	ticket, err := s.uploadTicketRepo.FindByObjectKey(ctx, storage.NormalizeObjectKey(objectKey))
	if err != nil {
		return err
	}
	if ticket == nil {
		return ErrUploadTokenInvalid
	}
	if ticket.ExpiresAt.Unix() <= s.now().Unix() {
		return ErrUploadTokenExpired
	}
	if claims != nil && ticket.DeviceID.String() != claims.DeviceID {
		return ErrUploadOwnershipMismatch
	}
	if deviceID != uuid.Nil && ticket.DeviceID != deviceID {
		return ErrUploadOwnershipMismatch
	}
	if ticket.ObjectKey != storage.NormalizeObjectKey(objectKey) {
		return ErrUploadTokenInvalid
	}
	if ticket.ContentType != storage.NormalizeContentType(contentType) {
		return ErrUploadTokenInvalid
	}
	if ticket.SizeBytes != sizeBytes {
		return ErrUploadTokenInvalid
	}
	return nil
}

type repositoryDevice struct {
	ID       uuid.UUID
	IsBanned bool
}
