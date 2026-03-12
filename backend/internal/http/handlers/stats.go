package handlers

import (
	"github.com/gofiber/fiber/v2"
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
	stats, err := h.svc.GetPublicStats(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch public stats")
	}

	c.Set("Cache-Control", "public, max-age=30, stale-while-revalidate=30")
	return response.OK(c, stats)
}
