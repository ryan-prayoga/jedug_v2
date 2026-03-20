package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/repository"
)

var (
	ErrDeviceBanned         = errors.New("device is banned")
	ErrCooldown             = errors.New("submission cooldown active")
	ErrLowTrustScore        = errors.New("device trust score too low")
	ErrMediaPersist         = errors.New("failed to persist submission media")
	ErrInvalidClientRequest = errors.New("invalid client_request_id")
	ErrIdempotencyConflict  = errors.New("client_request_id already used for a different report payload")
)

const (
	submitCooldown = 2 * time.Minute
	minTrustScore  = -10
)

type ReportService interface {
	SubmitReport(ctx context.Context, req SubmitReportRequest) (*SubmitReportResult, error)
}

type MediaInput struct {
	ObjectKey   string
	MimeType    string
	SizeBytes   int
	UploadToken string
	Width       *int
	Height      *int
	SHA256      *string
	IsPrimary   bool
	SortOrder   int
}

type SubmitReportRequest struct {
	ClientRequestID *string
	AnonToken       string
	ActorFollowerID *string
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
	deviceRepo         repository.DeviceRepository
	reportRepo         repository.ReportRepository
	locationNormalizer ReportLocationNormalizer
	uploadSvc          UploadService
}

func NewReportService(
	deviceRepo repository.DeviceRepository,
	reportRepo repository.ReportRepository,
	locationNormalizer ReportLocationNormalizer,
	uploadSvc UploadService,
) ReportService {
	return &reportService{
		deviceRepo:         deviceRepo,
		reportRepo:         reportRepo,
		locationNormalizer: locationNormalizer,
		uploadSvc:          uploadSvc,
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

	clientRequestID, clientRequestProvided, err := resolveClientRequestID(req.ClientRequestID)
	if err != nil {
		return nil, ErrInvalidClientRequest
	}
	requestFingerprint, err := buildReportRequestFingerprint(req)
	if err != nil {
		return nil, err
	}
	if clientRequestProvided {
		existing, findErr := s.reportRepo.FindByClientRequestID(ctx, device.ID, clientRequestID)
		if findErr != nil {
			return nil, findErr
		}
		if existing != nil {
			if existing.RequestFingerprint != "" && existing.RequestFingerprint != requestFingerprint {
				log.Printf("[REPORT] idempotency_conflict device=%s client_request_id=%s", device.ID, clientRequestID)
				return nil, ErrIdempotencyConflict
			}
			log.Printf("[ANTISPAM] idempotent_return device=%s client_request_id=%s", device.ID, clientRequestID)
			return &SubmitReportResult{
				IssueID:      existing.IssueID,
				SubmissionID: existing.SubmissionID,
				IsNewIssue:   existing.IsNewIssue,
			}, nil
		}
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

	// Cooldown check only applies to genuinely new submissions.
	lastSubmit, err := s.deviceRepo.FindLastSubmissionTime(ctx, device.ID)
	if err != nil {
		return nil, err
	}
	if lastSubmit != nil && time.Since(*lastSubmit) < submitCooldown {
		log.Printf("[ANTISPAM] cooldown device=%s last_submit=%s", device.ID, lastSubmit.Format(time.RFC3339))
		return nil, ErrCooldown
	}

	locationInfo := ReportLocationNormalization{
		RoadName:     nil,
		RegionName:   nil,
		CityName:     nil,
		DistrictName: nil,
		RegencyName:  nil,
		ProvinceName: nil,
	}
	if s.locationNormalizer != nil {
		locationInfo = s.locationNormalizer.NormalizeForReport(ctx, req.Longitude, req.Latitude)
	}

	mediaInputs := make([]repository.SubmitMediaInput, len(req.Media))
	mediaProofs := make([]ReportMediaProof, len(req.Media))
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
		mediaProofs[i] = ReportMediaProof{
			ObjectKey:   m.ObjectKey,
			MimeType:    m.MimeType,
			SizeBytes:   m.SizeBytes,
			UploadToken: m.UploadToken,
		}
	}

	if s.uploadSvc != nil {
		if err := s.uploadSvc.ValidateReportMedia(ctx, device.ID, mediaProofs); err != nil {
			return nil, err
		}
	}

	var actorFollowerID *uuid.UUID
	if req.ActorFollowerID != nil {
		trimmed := strings.TrimSpace(*req.ActorFollowerID)
		if trimmed != "" {
			parsedFollowerID, parseErr := uuid.Parse(trimmed)
			if parseErr != nil {
				return nil, errors.New("invalid actor_follower_id format")
			}
			actorFollowerID = &parsedFollowerID
		}
	}

	result, err := s.reportRepo.SubmitReport(ctx, repository.SubmitInput{
		ClientRequestID:    clientRequestID,
		DeviceID:           device.ID,
		ActorFollowerID:    actorFollowerID,
		PreferredRegionID:  locationInfo.RegionID,
		Longitude:          req.Longitude,
		Latitude:           req.Latitude,
		GPSAccuracyM:       req.GPSAccuracyM,
		CapturedAt:         req.CapturedAt,
		Severity:           req.Severity,
		HasCasualty:        req.HasCasualty,
		CasualtyCount:      req.CasualtyCount,
		Note:               req.Note,
		RoadName:           locationInfo.RoadName,
		DistrictName:       locationInfo.DistrictName,
		RegencyName:        locationInfo.RegencyName,
		ProvinceName:       locationInfo.ProvinceName,
		RequestFingerprint: requestFingerprint,
		Media:              mediaInputs,
	})
	if err != nil {
		log.Printf("[REPORT] repo_submit_failed device=%s severity=%d error=%v",
			device.ID, req.Severity, err)
		if errors.Is(err, repository.ErrSubmissionMediaPersistFailed) {
			return nil, ErrMediaPersist
		}
		if errors.Is(err, repository.ErrReportIdempotencyConflict) {
			return nil, ErrIdempotencyConflict
		}
		return nil, err
	}

	return &SubmitReportResult{
		IssueID:      result.IssueID,
		SubmissionID: result.SubmissionID,
		IsNewIssue:   result.IsNewIssue,
	}, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func resolveClientRequestID(raw *string) (uuid.UUID, bool, error) {
	if raw == nil {
		return uuid.New(), false, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return uuid.New(), false, nil
	}
	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return uuid.Nil, false, err
	}
	return parsed, true, nil
}

type reportRequestFingerprint struct {
	Latitude      float64                         `json:"latitude"`
	Longitude     float64                         `json:"longitude"`
	Severity      int                             `json:"severity"`
	HasCasualty   bool                            `json:"has_casualty"`
	CasualtyCount int                             `json:"casualty_count"`
	Note          string                          `json:"note"`
	Media         []reportRequestFingerprintMedia `json:"media"`
}

type reportRequestFingerprintMedia struct {
	MimeType  string  `json:"mime_type"`
	SizeBytes int     `json:"size_bytes"`
	Width     *int    `json:"width,omitempty"`
	Height    *int    `json:"height,omitempty"`
	SHA256    *string `json:"sha256,omitempty"`
	IsPrimary bool    `json:"is_primary"`
	SortOrder int     `json:"sort_order"`
}

func buildReportRequestFingerprint(req SubmitReportRequest) (string, error) {
	payload := reportRequestFingerprint{
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		Severity:      req.Severity,
		HasCasualty:   req.HasCasualty,
		CasualtyCount: normalizedCasualtyCount(req.HasCasualty, req.CasualtyCount),
		Note:          strings.TrimSpace(valueOrEmpty(req.Note)),
		Media:         make([]reportRequestFingerprintMedia, len(req.Media)),
	}
	for i, media := range req.Media {
		payload.Media[i] = reportRequestFingerprintMedia{
			MimeType:  media.MimeType,
			SizeBytes: media.SizeBytes,
			Width:     media.Width,
			Height:    media.Height,
			SHA256:    media.SHA256,
			IsPrimary: media.IsPrimary,
			SortOrder: media.SortOrder,
		}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func normalizedCasualtyCount(hasCasualty bool, casualtyCount int) int {
	if !hasCasualty {
		return 0
	}
	return casualtyCount
}
