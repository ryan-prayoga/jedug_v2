package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
	"jedug_backend/internal/storage"
)

type AdminHandler struct {
	svc     service.AdminService
	storage storage.Driver
}

func NewAdminHandler(svc service.AdminService, s storage.Driver) *AdminHandler {
	return &AdminHandler{svc: svc, storage: s}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AdminHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	if req.Username == "" || req.Password == "" {
		return response.Error(c, fiber.StatusBadRequest, "username and password required")
	}

	token, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return response.Error(c, fiber.StatusUnauthorized, "invalid credentials")
		}
		return response.Error(c, fiber.StatusInternalServerError, "login failed")
	}

	return response.OK(c, fiber.Map{"token": token})
}

func (h *AdminHandler) Me(c *fiber.Ctx) error {
	username := c.Locals("admin_username").(string)
	return response.OK(c, fiber.Map{"username": username})
}

func (h *AdminHandler) ListIssues(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	issues, err := h.svc.ListIssues(c.Context(), limit, offset, status)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch issues")
	}

	return response.OK(c, issues)
}

func (h *AdminHandler) GetIssue(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid issue id")
	}

	detail, err := h.svc.GetIssueDetail(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch issue")
	}
	if detail == nil {
		return response.Error(c, fiber.StatusNotFound, "issue not found")
	}

	for _, m := range detail.Media {
		m.PublicURL = h.storage.BuildPublicURL(m.ObjectKey)
	}

	return response.OK(c, detail)
}

type moderationRequest struct {
	Reason *string `json:"reason"`
}

func (h *AdminHandler) HideIssue(c *fiber.Ctx) error {
	return h.moderateIssue(c, "hide")
}

func (h *AdminHandler) FixIssue(c *fiber.Ctx) error {
	return h.moderateIssue(c, "fix")
}

func (h *AdminHandler) RejectIssue(c *fiber.Ctx) error {
	return h.moderateIssue(c, "reject")
}

func (h *AdminHandler) UnhideIssue(c *fiber.Ctx) error {
	return h.moderateIssue(c, "unhide")
}

func (h *AdminHandler) moderateIssue(c *fiber.Ctx, action string) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid issue id")
	}

	var req moderationRequest
	_ = c.BodyParser(&req) // reason is optional

	username := c.Locals("admin_username").(string)

	switch action {
	case "hide":
		err = h.svc.HideIssue(c.Context(), id, username, req.Reason)
	case "fix":
		err = h.svc.FixIssue(c.Context(), id, username, req.Reason)
	case "reject":
		err = h.svc.RejectIssue(c.Context(), id, username, req.Reason)
	case "unhide":
		err = h.svc.UnhideIssue(c.Context(), id, username, req.Reason)
	}

	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "moderation action failed")
	}

	return response.OKMessage(c, action+" successful")
}

func (h *AdminHandler) BanDevice(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid device id")
	}

	var req moderationRequest
	_ = c.BodyParser(&req)

	username := c.Locals("admin_username").(string)

	if err := h.svc.BanDevice(c.Context(), id, username, req.Reason); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "ban device failed")
	}

	return response.OKMessage(c, "device banned")
}
