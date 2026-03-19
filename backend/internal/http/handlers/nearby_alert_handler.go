package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

type NearbyAlertHandler struct {
	svc     service.NearbyAlertService
	authSvc service.FollowerAuthService
}

type nearbyAlertCreateBody struct {
	FollowerToken string  `json:"follower_token"`
	Label         *string `json:"label"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	RadiusM       int     `json:"radius_m"`
	Enabled       *bool   `json:"enabled"`
}

type nearbyAlertPatchBody struct {
	FollowerToken string   `json:"follower_token"`
	Label         *string  `json:"label"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	RadiusM       *int     `json:"radius_m"`
	Enabled       *bool    `json:"enabled"`
}

type nearbyAlertDeleteBody struct {
	FollowerToken string `json:"follower_token"`
}

func NewNearbyAlertHandler(svc service.NearbyAlertService, authSvc service.FollowerAuthService) *NearbyAlertHandler {
	return &NearbyAlertHandler{svc: svc, authSvc: authSvc}
}

func (h *NearbyAlertHandler) List(c *fiber.Ctx) error {
	followerID, err := authenticateFollowerToken(c, h.authSvc)
	if err != nil {
		return mapFollowerAuthError(c, err)
	}

	items, err := h.svc.List(c.Context(), followerID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch nearby alerts")
	}

	return response.OK(c, items)
}

func (h *NearbyAlertHandler) Create(c *fiber.Ctx) error {
	var body nearbyAlertCreateBody
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	followerID, err := authenticateFollowerTokenWithBody(c, h.authSvc, body.FollowerToken)
	if err != nil {
		return mapFollowerAuthError(c, err)
	}

	item, svcErr := h.svc.Create(c.Context(), followerID, service.NearbyAlertCreateInput{
		Label:     normalizeNearbyAlertBodyLabel(body.Label),
		Latitude:  body.Latitude,
		Longitude: body.Longitude,
		RadiusM:   body.RadiusM,
		Enabled:   body.Enabled,
	})
	if svcErr != nil {
		return mapNearbyAlertError(c, svcErr)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    item,
	})
}

func (h *NearbyAlertHandler) Patch(c *fiber.Ctx) error {
	subscriptionID, err := parseNearbyAlertSubscriptionID(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	var body nearbyAlertPatchBody
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	followerID, err := authenticateFollowerTokenWithBody(c, h.authSvc, body.FollowerToken)
	if err != nil {
		return mapFollowerAuthError(c, err)
	}

	item, svcErr := h.svc.Update(c.Context(), followerID, subscriptionID, domain.NearbyAlertSubscriptionPatch{
		Label:     normalizeNearbyAlertBodyLabel(body.Label),
		Latitude:  body.Latitude,
		Longitude: body.Longitude,
		RadiusM:   body.RadiusM,
		Enabled:   body.Enabled,
	})
	if svcErr != nil {
		return mapNearbyAlertError(c, svcErr)
	}

	return response.OK(c, item)
}

func (h *NearbyAlertHandler) Delete(c *fiber.Ctx) error {
	subscriptionID, err := parseNearbyAlertSubscriptionID(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	var body nearbyAlertDeleteBody
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&body); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid request body")
		}
	}

	followerID, err := authenticateFollowerTokenWithBody(c, h.authSvc, body.FollowerToken)
	if err != nil {
		return mapFollowerAuthError(c, err)
	}

	deleted, svcErr := h.svc.Delete(c.Context(), followerID, subscriptionID)
	if svcErr != nil {
		return mapNearbyAlertError(c, svcErr)
	}

	return response.OK(c, fiber.Map{"deleted": deleted})
}

func mapNearbyAlertError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrNearbyAlertInvalidCoordinates), errors.Is(err, service.ErrNearbyAlertCoordinatePairRequired), errors.Is(err, service.ErrNearbyAlertInvalidRadius), errors.Is(err, service.ErrNearbyAlertLabelTooLong), errors.Is(err, service.ErrNearbyAlertPatchRequired):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrNearbyAlertLimitExceeded):
		return response.ErrorWithCode(c, fiber.StatusTooManyRequests, "nearby_alert_limit_exceeded", "Batas lokasi pantauan untuk browser ini sudah tercapai")
	case errors.Is(err, service.ErrNearbyAlertNotFound):
		return response.Error(c, fiber.StatusNotFound, "nearby alert subscription not found")
	default:
		return response.Error(c, fiber.StatusInternalServerError, "failed to process nearby alert subscription")
	}
}

func parseNearbyAlertSubscriptionID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, errors.New("nearby alert id must be a valid uuid")
	}
	return id, nil
}

func normalizeNearbyAlertBodyLabel(label *string) *string {
	if label == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*label)
	return &trimmed
}
