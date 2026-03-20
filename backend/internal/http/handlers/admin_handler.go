package handlers

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
	"jedug_backend/internal/storage"
)

type AdminHandler struct {
	svc     service.AdminService
	storage *storage.Service
	secure  bool
}

func NewAdminHandler(svc service.AdminService, s *storage.Service, secure bool) *AdminHandler {
	return &AdminHandler{svc: svc, storage: s, secure: secure}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AdminHandler) Login(c *fiber.Ctx) error {
	rid := requestID(c)

	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ADMIN] login_parse_failed rid=%s ip=%s err=%v", rid, c.IP(), err)
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	if req.Username == "" || req.Password == "" {
		log.Printf("[ADMIN] login_validation_failed rid=%s ip=%s", rid, c.IP())
		return response.Error(c, fiber.StatusBadRequest, "username and password required")
	}

	fingerprint := adminLoginFingerprint(c.IP(), req.Username)
	token, err := h.svc.Login(req.Username, req.Password, fingerprint)
	if err != nil {
		var throttledErr *service.AdminLoginThrottleError
		if errors.As(err, &throttledErr) {
			log.Printf("[ADMIN] login_throttled rid=%s ip=%s username=%s retry_after=%s", rid, c.IP(), strings.TrimSpace(req.Username), throttledErr.RetryAfter)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success":     false,
				"message":     "Terlalu banyak percobaan login. Coba lagi nanti.",
				"retry_after": maxRetryAfterSeconds(throttledErr.RetryAfter),
			})
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			log.Printf("[ADMIN] login_failed rid=%s ip=%s username=%s", rid, c.IP(), strings.TrimSpace(req.Username))
			return response.Error(c, fiber.StatusUnauthorized, "username atau password salah")
		}
		log.Printf("[ADMIN] login_internal_error rid=%s ip=%s username=%s err=%v", rid, c.IP(), strings.TrimSpace(req.Username), err)
		return response.Error(c, fiber.StatusInternalServerError, "login failed")
	}

	h.setSessionCookie(c, token, service.AdminSessionTTL)
	log.Printf("[ADMIN] login_success rid=%s ip=%s username=%s", rid, c.IP(), strings.TrimSpace(req.Username))
	return response.OK(c, fiber.Map{
		"username": strings.TrimSpace(req.Username),
	})
}

func (h *AdminHandler) Logout(c *fiber.Ctx) error {
	token, _ := c.Locals("admin_session_token").(string)
	username, _ := c.Locals("admin_username").(string)
	h.svc.RevokeSession(token)
	h.clearSessionCookie(c)
	log.Printf("[ADMIN] logout rid=%s username=%s", requestID(c), strings.TrimSpace(username))
	return response.OKMessage(c, "logout successful")
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
		m.PublicURL = h.storage.ResolvePublicURL(m.ObjectKey)
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
	rid := requestID(c)
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
		if errors.Is(err, service.ErrModerationTargetNotFound) {
			log.Printf("[ADMIN] moderation_not_found rid=%s action=%s issue=%s admin=%s", rid, action, id, username)
			return response.Error(c, fiber.StatusNotFound, "issue not found")
		}
		log.Printf("[ADMIN] moderation_failed rid=%s action=%s issue=%s admin=%s err=%v", rid, action, id, username, err)
		return response.Error(c, fiber.StatusInternalServerError, "moderation action failed")
	}

	log.Printf("[ADMIN] moderation_success rid=%s action=%s issue=%s admin=%s", rid, action, id, username)
	return response.OKMessage(c, action+" successful")
}

func (h *AdminHandler) BanDevice(c *fiber.Ctx) error {
	rid := requestID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid device id")
	}

	var req moderationRequest
	_ = c.BodyParser(&req)

	username := c.Locals("admin_username").(string)

	if err := h.svc.BanDevice(c.Context(), id, username, req.Reason); err != nil {
		if errors.Is(err, service.ErrModerationTargetNotFound) {
			log.Printf("[ADMIN] ban_not_found rid=%s device=%s admin=%s", rid, id, username)
			return response.Error(c, fiber.StatusNotFound, "device not found")
		}
		log.Printf("[ADMIN] ban_failed rid=%s device=%s admin=%s err=%v", rid, id, username, err)
		return response.Error(c, fiber.StatusInternalServerError, "ban device failed")
	}

	log.Printf("[ADMIN] ban_success rid=%s device=%s admin=%s", rid, id, username)
	return response.OKMessage(c, "device banned")
}

func (h *AdminHandler) setSessionCookie(c *fiber.Ctx, token string, ttl time.Duration) {
	c.Cookie(&fiber.Cookie{
		Name:     service.AdminSessionCookieName,
		Value:    token,
		Path:     service.AdminSessionCookiePath,
		HTTPOnly: true,
		Secure:   h.secure,
		SameSite: fiber.CookieSameSiteStrictMode,
		MaxAge:   int(ttl.Seconds()),
		Expires:  time.Now().Add(ttl),
	})
}

func (h *AdminHandler) clearSessionCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     service.AdminSessionCookieName,
		Value:    "",
		Path:     service.AdminSessionCookiePath,
		HTTPOnly: true,
		Secure:   h.secure,
		SameSite: fiber.CookieSameSiteStrictMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func adminLoginFingerprint(ip, username string) string {
	return strings.TrimSpace(ip) + "|" + strings.ToLower(strings.TrimSpace(username))
}

func maxRetryAfterSeconds(duration time.Duration) int {
	seconds := int(duration.Seconds())
	if seconds < 1 {
		return 1
	}
	return seconds
}
