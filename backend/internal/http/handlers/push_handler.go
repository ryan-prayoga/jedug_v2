package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

type PushHandler struct {
	svc service.PushService
}

type pushSubscriptionBody struct {
	FollowerID   string `json:"follower_id"`
	Endpoint     string `json:"endpoint"`
	Subscription struct {
		Endpoint string `json:"endpoint"`
		Keys     struct {
			P256DH string `json:"p256dh"`
			Auth   string `json:"auth"`
		} `json:"keys"`
	} `json:"subscription"`
}

type pushUnsubscribeBody struct {
	FollowerID string `json:"follower_id"`
	Endpoint   string `json:"endpoint"`
}

func NewPushHandler(svc service.PushService) *PushHandler {
	return &PushHandler{svc: svc}
}

func (h *PushHandler) Status(c *fiber.Ctx) error {
	followerID, err := uuid.Parse(c.Query("follower_id"))
	if err != nil || followerID == uuid.Nil {
		return response.Error(c, fiber.StatusBadRequest, "follower_id must be a valid UUID")
	}

	status, err := h.svc.GetStatus(c.Context(), followerID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch push status")
	}

	return response.OK(c, status)
}

func (h *PushHandler) Subscribe(c *fiber.Ctx) error {
	var body pushSubscriptionBody
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	followerID, err := parsePushFollowerID(body.FollowerID)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	endpoint := firstNonEmpty(body.Subscription.Endpoint, body.Endpoint)
	status, svcErr := h.svc.Subscribe(c.Context(), service.PushSubscribeInput{
		FollowerID: followerID,
		Endpoint:   endpoint,
		P256DH:     body.Subscription.Keys.P256DH,
		Auth:       body.Subscription.Keys.Auth,
		UserAgent:  optionalHeader(c.Get("User-Agent")),
	})
	if svcErr != nil {
		return mapPushError(c, svcErr)
	}

	return response.OK(c, status)
}

func (h *PushHandler) Unsubscribe(c *fiber.Ctx) error {
	var body pushUnsubscribeBody
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&body); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid request body")
		}
	}

	followerID, err := parsePushFollowerID(firstNonEmpty(body.FollowerID, c.Query("follower_id")))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	status, unsubscribed, svcErr := h.svc.Unsubscribe(c.Context(), service.PushUnsubscribeInput{
		FollowerID: followerID,
		Endpoint:   firstNonEmpty(body.Endpoint, c.Query("endpoint")),
	})
	if svcErr != nil {
		return mapPushError(c, svcErr)
	}

	return response.OK(c, fiber.Map{
		"enabled":            status.Enabled,
		"subscribed":         status.Subscribed,
		"subscription_count": status.SubscriptionCount,
		"unsubscribed":       unsubscribed,
	})
}

func parsePushFollowerID(raw string) (uuid.UUID, error) {
	followerID, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || followerID == uuid.Nil {
		return uuid.Nil, errors.New("follower_id must be a valid UUID")
	}
	return followerID, nil
}

func mapPushError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrPushDisabled):
		return response.Error(c, fiber.StatusServiceUnavailable, "browser push notification belum dikonfigurasi")
	case err != nil && strings.Contains(err.Error(), "required"):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	default:
		return response.Error(c, fiber.StatusInternalServerError, "failed to process push subscription")
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func optionalHeader(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
