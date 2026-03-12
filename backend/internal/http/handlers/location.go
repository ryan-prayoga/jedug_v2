package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

type LocationHandler struct {
	svc service.LocationService
}

func NewLocationHandler(svc service.LocationService) *LocationHandler {
	return &LocationHandler{svc: svc}
}

func (h *LocationHandler) ResolveLabel(c *fiber.Ctx) error {
	latitude, err := parseCoordinate(c.Query("latitude"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "latitude must be a valid number")
	}
	if latitude < -90 || latitude > 90 {
		return response.Error(c, fiber.StatusBadRequest, "latitude must be between -90 and 90")
	}

	longitude, err := parseCoordinate(c.Query("longitude"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "longitude must be a valid number")
	}
	if longitude < -180 || longitude > 180 {
		return response.Error(c, fiber.StatusBadRequest, "longitude must be between -180 and 180")
	}

	log.Printf("[LOCATION_LABEL] request_received lat=%.6f lon=%.6f", latitude, longitude)

	result, err := h.svc.ResolveLabel(c.Context(), longitude, latitude)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to resolve location label")
	}

	log.Printf(
		"[LOCATION_LABEL] response_ready lat=%.6f lon=%.6f source=%s label=%q region_name=%q",
		latitude,
		longitude,
		result.Source,
		valueOrNilString(result.Label),
		valueOrNilString(result.RegionName),
	)

	return response.OK(c, result)
}

func parseCoordinate(raw string) (float64, error) {
	return strconv.ParseFloat(raw, 64)
}

func valueOrNilString(value *string) string {
	if value == nil {
		return "<nil>"
	}
	return *value
}
