package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

type DeviceHandler struct {
	svc service.DeviceService
}

func NewDeviceHandler(svc service.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

func (h *DeviceHandler) Bootstrap(c *fiber.Ctx) error {
	anonToken := c.Get("X-Device-Token")
	userAgent := c.Get("User-Agent")
	ipAddress := c.IP()

	req := service.BootstrapRequest{
		UserAgent: strPtr(userAgent),
		IPAddress: strPtr(ipAddress),
	}
	if anonToken != "" {
		req.AnonToken = &anonToken
	}

	result, err := h.svc.Bootstrap(c.Context(), req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to bootstrap device")
	}

	return response.OK(c, result)
}

type consentBody struct {
	AnonToken      string `json:"anon_token"`
	TermsVersion   string `json:"terms_version"`
	PrivacyVersion string `json:"privacy_version"`
}

func (h *DeviceHandler) Consent(c *fiber.Ctx) error {
	var body consentBody
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	if body.AnonToken == "" {
		return response.Error(c, fiber.StatusBadRequest, "anon_token is required")
	}
	if body.TermsVersion == "" || body.PrivacyVersion == "" {
		return response.Error(c, fiber.StatusBadRequest, "terms_version and privacy_version are required")
	}

	err := h.svc.RecordConsent(c.Context(), service.ConsentRequest{
		AnonToken:      body.AnonToken,
		TermsVersion:   body.TermsVersion,
		PrivacyVersion: body.PrivacyVersion,
		IPAddress:      strPtr(c.IP()),
		UserAgent:      strPtr(c.Get("User-Agent")),
	})
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			return response.Error(c, fiber.StatusNotFound, "device not found")
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to record consent")
	}

	return response.OKMessage(c, "consent recorded")
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
