package handlers

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
	"jedug_backend/internal/storage"
)

type ReportHandler struct {
	svc service.ReportService
}

func NewReportHandler(svc service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

type reportMediaInput struct {
	ObjectKey string  `json:"object_key"`
	MimeType  string  `json:"mime_type"`
	SizeBytes int     `json:"size_bytes"`
	Width     *int    `json:"width"`
	Height    *int    `json:"height"`
	SHA256    *string `json:"sha256"`
	IsPrimary bool    `json:"is_primary"`
}

type submitReportBody struct {
	ClientRequestID *string            `json:"client_request_id"`
	AnonToken       string             `json:"anon_token"`
	Latitude        float64            `json:"latitude"`
	Longitude       float64            `json:"longitude"`
	GPSAccuracyM    *float64           `json:"gps_accuracy_m"`
	Severity        int                `json:"severity"`
	Note            *string            `json:"note"`
	HasCasualty     bool               `json:"has_casualty"`
	CasualtyCount   int                `json:"casualty_count"`
	CapturedAt      *time.Time         `json:"captured_at"`
	Media           []reportMediaInput `json:"media"`
}

func (h *ReportHandler) Submit(c *fiber.Ctx) error {
	var body submitReportBody
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	if err := validateReportBody(&body); err != nil {
		log.Printf("[ANTISPAM] validation_failed ip=%s reason=%s", c.IP(), err.Error())
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	media := make([]service.MediaInput, len(body.Media))
	for i, m := range body.Media {
		media[i] = service.MediaInput{
			ObjectKey: m.ObjectKey,
			MimeType:  m.MimeType,
			SizeBytes: m.SizeBytes,
			Width:     m.Width,
			Height:    m.Height,
			SHA256:    m.SHA256,
			IsPrimary: m.IsPrimary,
			SortOrder: i,
		}
	}

	result, err := h.svc.SubmitReport(c.Context(), service.SubmitReportRequest{
		ClientRequestID: body.ClientRequestID,
		AnonToken:       body.AnonToken,
		Latitude:        body.Latitude,
		Longitude:       body.Longitude,
		GPSAccuracyM:    body.GPSAccuracyM,
		CapturedAt:      body.CapturedAt,
		Severity:        body.Severity,
		HasCasualty:     body.HasCasualty,
		CasualtyCount:   body.CasualtyCount,
		Note:            body.Note,
		Media:           media,
	})
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			return response.Error(c, fiber.StatusUnauthorized, "device not found; bootstrap first")
		}
		if errors.Is(err, service.ErrDeviceBanned) {
			log.Printf("[ANTISPAM] banned_submit ip=%s", c.IP())
			return response.Error(c, fiber.StatusForbidden, "device is banned")
		}
		if errors.Is(err, service.ErrCooldown) {
			log.Printf("[ANTISPAM] cooldown_submit ip=%s", c.IP())
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success":     false,
				"message":     "Tunggu beberapa menit sebelum mengirim laporan berikutnya.",
				"retry_after": 120,
			})
		}
		if errors.Is(err, service.ErrLowTrustScore) {
			log.Printf("[ANTISPAM] low_trust_submit ip=%s", c.IP())
			return response.Error(c, fiber.StatusForbidden, "Akun tidak diizinkan mengirim laporan saat ini.")
		}
		log.Printf("[REPORT] submit_internal_error ip=%s error=%v", c.IP(), err)
		return response.Error(c, fiber.StatusInternalServerError, "failed to submit report")
	}

	status := fiber.StatusOK
	msg := "report submitted"
	if result.IsNewIssue {
		status = fiber.StatusCreated
		msg = "report submitted, new issue created"
	}

	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": msg,
		"data":    result,
	})
}

func validateReportBody(b *submitReportBody) error {
	if b.AnonToken == "" {
		return errors.New("anon_token is required")
	}
	if b.Latitude < -90 || b.Latitude > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if b.Longitude < -180 || b.Longitude > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	if b.Severity < 1 || b.Severity > 5 {
		return errors.New("severity must be between 1 and 5")
	}
	if len(b.Media) == 0 {
		return errors.New("at least one media item is required")
	}
	if len(b.Media) > 5 {
		return errors.New("maximum 5 media items allowed")
	}
	if b.CasualtyCount < 0 {
		return errors.New("casualty_count must be >= 0")
	}
	if !b.HasCasualty && b.CasualtyCount > 0 {
		b.CasualtyCount = 0
	}
	if b.Note != nil && len(*b.Note) > 500 {
		return errors.New("note must not exceed 500 characters")
	}
	for i, m := range b.Media {
		if m.ObjectKey == "" {
			return errors.New("media[" + itoa(i) + "].object_key is required")
		}
		if err := storage.ValidateSubmittedMedia(m.ObjectKey, m.MimeType, m.SizeBytes); err != nil {
			return errors.New("media[" + itoa(i) + "]: " + err.Error())
		}
	}
	return nil
}

func itoa(n int) string { return strconv.Itoa(n) }
