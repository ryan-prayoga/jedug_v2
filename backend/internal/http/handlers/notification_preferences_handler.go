package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

type NotificationPreferencesHandler struct {
	svc     service.NotificationPreferencesService
	authSvc service.FollowerAuthService
}

type notificationPreferencesPatchBody struct {
	FollowerToken            string `json:"follower_token"`
	NotificationsEnabled     *bool  `json:"notifications_enabled"`
	InAppEnabled             *bool  `json:"in_app_enabled"`
	PushEnabled              *bool  `json:"push_enabled"`
	NotifyOnPhotoAdded       *bool  `json:"notify_on_photo_added"`
	NotifyOnStatusUpdated    *bool  `json:"notify_on_status_updated"`
	NotifyOnSeverityChanged  *bool  `json:"notify_on_severity_changed"`
	NotifyOnCasualtyReported *bool  `json:"notify_on_casualty_reported"`
	NotifyOnNearbyIssueCreated *bool `json:"notify_on_nearby_issue_created"`
}

func NewNotificationPreferencesHandler(svc service.NotificationPreferencesService, authSvc service.FollowerAuthService) *NotificationPreferencesHandler {
	return &NotificationPreferencesHandler{svc: svc, authSvc: authSvc}
}

func (h *NotificationPreferencesHandler) Get(c *fiber.Ctx) error {
	followerID, err := authenticateFollowerToken(c, h.authSvc)
	if err != nil {
		return mapFollowerAuthError(c, err)
	}

	prefs, err := h.svc.GetByFollowerID(c.Context(), followerID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch notification preferences")
	}

	return response.OK(c, prefs)
}

func (h *NotificationPreferencesHandler) Patch(c *fiber.Ctx) error {
	var body notificationPreferencesPatchBody
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	rawToken, err := parseFollowerToken(firstNonEmpty(
		strings.TrimSpace(body.FollowerToken),
		c.Query("follower_token"),
		c.Get("X-Follower-Token"),
	))
	if err != nil {
		return mapFollowerAuthError(c, err)
	}

	followerID, err := h.authSvc.Authenticate(c.Context(), rawToken)
	if err != nil {
		return mapFollowerAuthError(c, err)
	}

	patch := domain.NotificationPreferencesPatch{
		NotificationsEnabled:     body.NotificationsEnabled,
		InAppEnabled:             body.InAppEnabled,
		PushEnabled:              body.PushEnabled,
		NotifyOnPhotoAdded:       body.NotifyOnPhotoAdded,
		NotifyOnStatusUpdated:    body.NotifyOnStatusUpdated,
		NotifyOnSeverityChanged:  body.NotifyOnSeverityChanged,
		NotifyOnCasualtyReported: body.NotifyOnCasualtyReported,
		NotifyOnNearbyIssueCreated: body.NotifyOnNearbyIssueCreated,
	}
	if patch.IsEmpty() {
		return response.Error(c, fiber.StatusBadRequest, "at least one notification preference must be provided")
	}

	prefs, err := h.svc.Update(c.Context(), followerID, patch)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to update notification preferences")
	}

	return response.OK(c, prefs)
}
