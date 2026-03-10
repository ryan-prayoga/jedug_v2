package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

type FlagHandler struct {
	svc service.FlagService
}

func NewFlagHandler(svc service.FlagService) *FlagHandler {
	return &FlagHandler{svc: svc}
}

type flagIssueBody struct {
	AnonToken string  `json:"anon_token"`
	Reason    string  `json:"reason"`
	Note      *string `json:"note"`
}

func (h *FlagHandler) FlagIssue(c *fiber.Ctx) error {
	issueID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid issue id")
	}

	var body flagIssueBody
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	if body.AnonToken == "" {
		return response.Error(c, fiber.StatusBadRequest, "anon_token is required")
	}
	if body.Reason == "" {
		return response.Error(c, fiber.StatusBadRequest, "reason is required")
	}

	err = h.svc.FlagIssue(c.Context(), service.FlagIssueRequest{
		AnonToken: body.AnonToken,
		IssueID:   issueID,
		Reason:    body.Reason,
		Note:      body.Note,
	})
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			return response.Error(c, fiber.StatusUnauthorized, "device not found")
		}
		if errors.Is(err, service.ErrDeviceBanned) {
			return response.Error(c, fiber.StatusForbidden, "device is banned")
		}
		if errors.Is(err, service.ErrAlreadyFlagged) {
			return response.OKMessage(c, "already flagged")
		}
		if errors.Is(err, service.ErrInvalidFlagReason) {
			return response.Error(c, fiber.StatusBadRequest, "invalid reason; allowed: spam, hoax, off_topic, duplicate, abuse, other")
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to flag issue")
	}

	return response.OKMessage(c, "issue flagged")
}
