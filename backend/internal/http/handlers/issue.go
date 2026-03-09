package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/repository"
	"jedug_backend/internal/service"
	"jedug_backend/internal/storage"
)

type IssueHandler struct {
	svc     service.IssueService
	storage storage.Driver
}

func NewIssueHandler(svc service.IssueService, s storage.Driver) *IssueHandler {
	return &IssueHandler{svc: svc, storage: s}
}

func (h *IssueHandler) List(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	var bbox *repository.BBoxFilter
	if raw := c.Query("bbox"); raw != "" {
		parts := strings.Split(raw, ",")
		if len(parts) != 4 {
			return response.Error(c, fiber.StatusBadRequest, "bbox must be minLng,minLat,maxLng,maxLat")
		}
		vals := make([]float64, 4)
		for i, p := range parts {
			v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
			if err != nil {
				return response.Error(c, fiber.StatusBadRequest, "bbox contains invalid value")
			}
			vals[i] = v
		}
		bbox = &repository.BBoxFilter{MinLng: vals[0], MinLat: vals[1], MaxLng: vals[2], MaxLat: vals[3]}
	}

	issues, err := h.svc.List(c.Context(), limit, offset, status, bbox)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch issues")
	}

	return response.OK(c, issues)
}

func (h *IssueHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid issue id")
	}

	detail, err := h.svc.GetByIDWithDetail(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch issue")
	}
	if detail == nil {
		return response.Error(c, fiber.StatusNotFound, "issue not found")
	}

	// Populate public_url for each media item using the storage driver
	for _, m := range detail.Media {
		m.PublicURL = h.storage.BuildPublicURL(m.ObjectKey)
	}

	return response.OK(c, detail)
}

