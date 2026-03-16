package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

type FollowerAuthHandler struct {
	svc service.FollowerAuthService
}

type followerAuthBody struct {
	FollowerID string `json:"follower_id"`
}

func NewFollowerAuthHandler(svc service.FollowerAuthService) *FollowerAuthHandler {
	return &FollowerAuthHandler{svc: svc}
}

func (h *FollowerAuthHandler) Issue(c *fiber.Ctx) error {
	var body followerAuthBody
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&body); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid request body")
		}
	}

	followerID, err := parseFollowerUUID(firstNonEmpty(body.FollowerID, c.Query("follower_id")))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	token, svcErr := h.svc.IssueForNotificationAccess(c.Context(), followerID, c.Get("X-Device-Token"))
	if svcErr != nil {
		return mapFollowerAuthError(c, svcErr)
	}

	return response.OK(c, token)
}

func mapFollowerAuthError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrDeviceBootstrapRequired):
		return response.ErrorWithCode(c, fiber.StatusUnauthorized, "device_bootstrap_required", "bootstrap device diperlukan sebelum mengakses notifikasi")
	case errors.Is(err, service.ErrFollowerBindingMismatch):
		return response.ErrorWithCode(c, fiber.StatusForbidden, "follower_binding_mismatch", "browser ini bukan pemilik follower tersebut")
	case errors.Is(err, service.ErrFollowerBindingNotFound):
		return response.ErrorWithCode(c, fiber.StatusForbidden, "follower_binding_not_found", "follower ini perlu diikat ulang dari browser yang sama")
	case errors.Is(err, service.ErrFollowerTokenRequired):
		return response.ErrorWithCode(c, fiber.StatusUnauthorized, "follower_token_required", "follower_token is required")
	case errors.Is(err, service.ErrFollowerTokenExpired):
		return response.ErrorWithCode(c, fiber.StatusUnauthorized, "follower_token_expired", "follower token sudah kedaluwarsa")
	case errors.Is(err, service.ErrFollowerTokenInvalid):
		return response.ErrorWithCode(c, fiber.StatusUnauthorized, "follower_token_invalid", "follower token tidak valid")
	default:
		return response.Error(c, fiber.StatusInternalServerError, "failed to authorize follower")
	}
}

func parseFollowerToken(raw string) (string, error) {
	token := firstNonEmpty(raw)
	if token == "" {
		return "", service.ErrFollowerTokenRequired
	}
	return token, nil
}

func authenticateFollowerToken(c *fiber.Ctx, authSvc service.FollowerAuthService) (uuid.UUID, error) {
	token, err := parseFollowerToken(firstNonEmpty(
		c.Query("follower_token"),
		c.Get("X-Follower-Token"),
	))
	if err != nil {
		return uuid.Nil, err
	}
	return authSvc.Authenticate(c.Context(), token)
}
