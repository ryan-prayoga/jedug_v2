package handlers

import (
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
	objectKey := c.Params("*")
	if err := storage.ValidateObjectKey(objectKey); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	body := c.Body()
	if len(body) == 0 {
		return response.Error(c, fiber.StatusBadRequest, "empty file body")
	}
	if len(body) > storage.MaxFileSizeBytes {
		return response.Error(c, fiber.StatusBadRequest, "file exceeds maximum size of 20MB")
	}

	if err := h.storage.Upload(c.Context(), objectKey, c.Get(fiber.HeaderContentType), body); err != nil {
		if storage.IsValidationError(err) {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to save file")
	}

	return response.OKMessage(c, "file uploaded")
}
