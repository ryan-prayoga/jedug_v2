package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

type IssueFollowHandler struct {
	svc     service.IssueFollowService
	authSvc service.FollowerAuthService
}

type issueFollowBody struct {
	FollowerID string `json:"follower_id"`
}

func NewIssueFollowHandler(svc service.IssueFollowService, authSvc service.FollowerAuthService) *IssueFollowHandler {
	return &IssueFollowHandler{svc: svc, authSvc: authSvc}
}

func (h *IssueFollowHandler) Follow(c *fiber.Ctx) error {
	issueID, followerID, err := parseFollowRequest(c)
	if err != nil {
		return err
	}

	authToken, authErr := h.authSvc.IssueForFollowMutation(c.Context(), issueID, followerID, c.Get("X-Device-Token"))
	if authErr != nil {
		return mapFollowerAuthError(c, authErr)
	}

	state, svcErr := h.svc.Follow(c.Context(), issueID, followerID)
	if svcErr != nil {
		return mapIssueFollowError(c, svcErr)
	}
	state.FollowerToken = authToken.Token
	state.FollowerTokenExpiresAt = &authToken.ExpiresAt

	return response.OK(c, state)
}

func (h *IssueFollowHandler) Unfollow(c *fiber.Ctx) error {
	issueID, followerID, err := parseFollowRequest(c)
	if err != nil {
		return err
	}

	authToken, authErr := h.authSvc.IssueForFollowMutation(c.Context(), issueID, followerID, c.Get("X-Device-Token"))
	if authErr != nil {
		return mapFollowerAuthError(c, authErr)
	}

	state, svcErr := h.svc.Unfollow(c.Context(), issueID, followerID)
	if svcErr != nil {
		return mapIssueFollowError(c, svcErr)
	}
	state.FollowerToken = authToken.Token
	state.FollowerTokenExpiresAt = &authToken.ExpiresAt

	return response.OK(c, state)
}

func (h *IssueFollowHandler) Count(c *fiber.Ctx) error {
	issueID, err := parseIssueID(c)
	if err != nil {
		return err
	}

	count, svcErr := h.svc.GetCount(c.Context(), issueID)
	if svcErr != nil {
		return mapIssueFollowError(c, svcErr)
	}

	return response.OK(c, count)
}

func (h *IssueFollowHandler) Status(c *fiber.Ctx) error {
	issueID, err := parseIssueID(c)
	if err != nil {
		return err
	}

	followerID, followerErr := parseFollowerUUID(c.Query("follower_id"))
	if followerErr != nil {
		return response.Error(c, fiber.StatusBadRequest, followerErr.Error())
	}

	authToken, authErr := h.authSvc.IssueForNotificationAccess(c.Context(), followerID, c.Get("X-Device-Token"))
	if authErr != nil {
		return mapFollowerAuthError(c, authErr)
	}

	state, svcErr := h.svc.GetStatus(c.Context(), issueID, followerID)
	if svcErr != nil {
		return mapIssueFollowError(c, svcErr)
	}
	state.FollowerToken = authToken.Token
	state.FollowerTokenExpiresAt = &authToken.ExpiresAt

	return response.OK(c, state)
}

func parseFollowRequest(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	issueID, err := parseIssueID(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	var body issueFollowBody
	if len(c.Body()) > 0 {
		if parseErr := c.BodyParser(&body); parseErr != nil {
			return uuid.Nil, uuid.Nil, response.Error(c, fiber.StatusBadRequest, "invalid request body")
		}
	}

	rawFollowerID := body.FollowerID
	if rawFollowerID == "" {
		rawFollowerID = c.Query("follower_id")
	}

	followerID, followerErr := parseFollowerUUID(rawFollowerID)
	if followerErr != nil {
		return uuid.Nil, uuid.Nil, response.Error(c, fiber.StatusBadRequest, followerErr.Error())
	}

	return issueID, followerID, nil
}

func parseIssueID(c *fiber.Ctx) (uuid.UUID, error) {
	issueID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, response.Error(c, fiber.StatusBadRequest, "invalid issue id")
	}
	return issueID, nil
}

func parseFollowerUUID(raw string) (uuid.UUID, error) {
	if raw == "" {
		return uuid.Nil, errors.New("follower_id is required")
	}

	followerID, err := uuid.Parse(raw)
	if err != nil || followerID == uuid.Nil {
		return uuid.Nil, errors.New("follower_id must be a valid uuid")
	}

	return followerID, nil
}

func mapIssueFollowError(c *fiber.Ctx, err error) error {
	if errors.Is(err, service.ErrIssueNotFound) {
		return response.Error(c, fiber.StatusNotFound, "issue not found")
	}

	return response.Error(c, fiber.StatusInternalServerError, "failed to process issue follow request")
}
