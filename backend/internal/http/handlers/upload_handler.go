package handlers

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/storage"
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/heic": true,
	"image/heif": true,
}

const maxFileSizeBytes = 20 * 1024 * 1024 // 20 MB

type UploadHandler struct {
	storage storage.Driver
}

func NewUploadHandler(s storage.Driver) *UploadHandler {
	return &UploadHandler{storage: s}
}

type presignRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int    `json:"size_bytes"`
}

// Presign generates an object key and upload target for the client.
// When STORAGE_DRIVER=local it returns an upload_url the client POSTs the raw file to.
// When STORAGE_DRIVER=r2 (future), it will return a presigned S3 URL instead.
func (h *UploadHandler) Presign(c *fiber.Ctx) error {
	var req presignRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	if req.Filename == "" {
		return response.Error(c, fiber.StatusBadRequest, "filename is required")
	}
	if !allowedMimeTypes[req.ContentType] {
		return response.Error(c, fiber.StatusBadRequest, "unsupported content type; allowed: jpeg, png, webp, heic, heif")
	}
	if req.SizeBytes <= 0 || req.SizeBytes > maxFileSizeBytes {
		return response.Error(c, fiber.StatusBadRequest, "size_bytes must be between 1 and 20971520")
	}

	objectKey := h.storage.GenerateObjectKey(req.Filename)

	return response.OK(c, fiber.Map{
		"object_key":  objectKey,
		"upload_mode": h.storage.UploadMode(),
		"upload_url":  h.storage.BuildUploadURL(objectKey),
		"public_url":  h.storage.BuildPublicURL(objectKey),
	})
}

// UploadFile accepts a raw file body and saves it to the local filesystem.
// This endpoint is only active when STORAGE_DRIVER=local.
// With R2, clients upload directly to the presigned URL — this endpoint becomes unused.
func (h *UploadHandler) UploadFile(c *fiber.Ctx) error {
	local, ok := h.storage.(*storage.LocalDriver)
	if !ok {
		return response.Error(c, fiber.StatusNotImplemented, "direct file upload is only available in local storage mode")
	}

	objectKey := c.Params("*")
	if objectKey == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing object key")
	}
	// Prevent path traversal
	if strings.Contains(objectKey, "..") {
		return response.Error(c, fiber.StatusBadRequest, "invalid object key")
	}

	body := c.Body()
	if len(body) == 0 {
		return response.Error(c, fiber.StatusBadRequest, "empty file body")
	}
	if len(body) > maxFileSizeBytes {
		return response.Error(c, fiber.StatusBadRequest, "file exceeds maximum size of 20MB")
	}

	absPath := local.AbsPath(objectKey)
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to prepare storage directory")
	}
	if err := os.WriteFile(absPath, body, 0644); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to save file")
	}

	return response.OKMessage(c, "file uploaded")
}
