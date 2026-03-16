package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

type StatsHandler struct {
	svc service.StatsService
}

func NewStatsHandler(svc service.StatsService) *StatsHandler {
	return &StatsHandler{svc: svc}
}

func (h *StatsHandler) Get(c *fiber.Ctx) error {
	provinceID, err := parseOptionalInt64(c.Query("province_id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "province_id must be a positive integer")
	}

	regencyID, err := parseOptionalInt64(c.Query("regency_id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "regency_id must be a positive integer")
	}

	stats, err := h.svc.GetPublicStats(c.Context(), domain.PublicStatsQuery{
		ProvinceID: provinceID,
		RegencyID:  regencyID,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch public stats")
	}

	c.Set("Cache-Control", "public, max-age=30, stale-while-revalidate=30")
	return response.OK(c, stats)
}

func (h *StatsHandler) GetRegionOptions(c *fiber.Ctx) error {
	options, err := h.svc.GetPublicRegionOptions(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch stats region options")
	}

	c.Set("Cache-Control", "public, max-age=60, stale-while-revalidate=60")
	return response.OK(c, options)
}

func parseOptionalInt64(raw string) (*int64, error) {
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, err
	}
	if value <= 0 {
		return nil, errors.New("value must be positive")
	}

	return &value, nil
}
