package handlers

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/storage"
)

type UploadHandler struct {
	storage *storage.Service
}

func NewUploadHandler(s *storage.Service) *UploadHandler {
	return &UploadHandler{storage: s}
}

type presignRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int    `json:"size_bytes"`
}

func (h *UploadHandler) Presign(c *fiber.Ctx) error {
	var req presignRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	result, err := h.storage.CreatePresign(c.Context(), storage.UploadRequest{
		Filename:    req.Filename,
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
	})
	if err != nil {
		if storage.IsValidationError(err) {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to create upload target")
	}

	return response.OK(c, result)
}

func (h *UploadHandler) UploadFile(c *fiber.Ctx) error {
	local := h.storage.LegacyLocal()
	if local == nil || h.storage.UploadMode() != "local" {
		return response.Error(c, fiber.StatusNotImplemented, "direct file upload is only available in local storage mode")
	}

	objectKey := c.Params("*")
	if err := storage.ValidateObjectKey(objectKey); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	body := c.Body()
	if len(body) == 0 {
		return response.Error(c, fiber.StatusBadRequest, "empty file body")
	}
	if err := storage.ValidateSubmittedMedia(objectKey, c.Get(fiber.HeaderContentType), len(body)); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}
	if len(body) > storage.MaxFileSizeBytes {
		return response.Error(c, fiber.StatusBadRequest, "file exceeds maximum size of 20MB")
	}

	absPath := local.AbsPath(objectKey)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to prepare storage directory")
	}
	if err := os.WriteFile(absPath, body, 0o644); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to save file")
	}

	return response.OKMessage(c, "file uploaded")
}
