package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// GET /api/v1/notifications?follower_id=<uuid>&limit=50
func (h *NotificationHandler) List(c *fiber.Ctx) error {
	followerID, err := uuid.Parse(c.Query("follower_id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "follower_id must be a valid UUID")
	}

	limit := c.QueryInt("limit", 50)
	notifications, err := h.svc.GetByFollowerID(c.Context(), followerID, limit)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch notifications")
	}

	return response.OK(c, fiber.Map{
		"items": notifications,
	})
}
