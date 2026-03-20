package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
	"jedug_backend/internal/storage"
)

type UploadHandler struct {
	svc     service.UploadService
	storage *storage.Service
}

func NewUploadHandler(svc service.UploadService, store *storage.Service) *UploadHandler {
	return &UploadHandler{svc: svc, storage: store}
}

type presignRequest struct {
	AnonToken   string `json:"anon_token"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int    `json:"size_bytes"`
}

func (h *UploadHandler) Presign(c *fiber.Ctx) error {
	var req presignRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	result, err := h.svc.CreateReportUpload(c.Context(), service.CreateReportUploadRequest{
		AnonToken:   req.AnonToken,
		Filename:    req.Filename,
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
	})
	if err != nil {
		if errors.Is(err, service.ErrUploadDeviceBootstrapNeeded) || errors.Is(err, service.ErrDeviceNotFound) {
			return response.ErrorWithCode(c, fiber.StatusUnauthorized, "DEVICE_NOT_READY", "device not found; bootstrap first")
		}
		if errors.Is(err, service.ErrDeviceBanned) {
			return response.ErrorWithCode(c, fiber.StatusForbidden, "DEVICE_BANNED", "device is banned")
		}
		if errors.Is(err, service.ErrUploadPendingLimitReached) {
			return response.ErrorWithCode(c, fiber.StatusTooManyRequests, "UPLOAD_RATE_LIMITED", "terlalu banyak upload yang belum dipakai. Coba lagi beberapa saat lagi.")
		}
		if storage.IsValidationError(err) {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to create upload target")
	}

	return response.OK(c, result)
}

func (h *UploadHandler) UploadFile(c *fiber.Ctx) error {
	objectKey := c.Params("*")

	body := c.Body()
	if len(body) == 0 {
		return response.Error(c, fiber.StatusBadRequest, "empty file body")
	}

	contentType := c.Get(fiber.HeaderContentType)
	uploadToken := c.Get("X-Upload-Token")
	if err := h.svc.ValidateLocalUpload(c.Context(), uploadToken, objectKey, contentType, len(body)); err != nil {
		switch {
		case errors.Is(err, service.ErrUploadTokenRequired):
			return response.ErrorWithCode(c, fiber.StatusUnauthorized, "UPLOAD_TOKEN_REQUIRED", "upload token is required")
		case errors.Is(err, service.ErrUploadTokenInvalid), errors.Is(err, service.ErrUploadTokenExpired):
			return response.ErrorWithCode(c, fiber.StatusForbidden, "UPLOAD_TOKEN_INVALID", "upload token is invalid or expired")
		case errors.Is(err, service.ErrUploadAlreadyUsed):
			return response.ErrorWithCode(c, fiber.StatusConflict, "MEDIA_ALREADY_USED", "media has already been attached to another report")
		case storage.IsValidationError(err):
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		default:
			return response.Error(c, fiber.StatusInternalServerError, "failed to validate upload")
		}
	}

	if err := h.storage.Upload(c.Context(), objectKey, contentType, body); err != nil {
		if storage.IsValidationError(err) {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to save file")
	}
	return response.OKMessage(c, "file uploaded")
}
