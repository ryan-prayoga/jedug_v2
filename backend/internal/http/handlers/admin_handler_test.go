package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

func TestAdminHandlerHideIssueReturns404ForMissingIssue(t *testing.T) {
	app := fiber.New()
	handler := NewAdminHandler(&adminServiceHTTPFake{hideErr: service.ErrModerationTargetNotFound}, nil, false)

	app.Post("/admin/issues/:id/hide", func(c *fiber.Ctx) error {
		c.Locals("admin_username", "admin")
		return handler.HideIssue(c)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/admin/issues/"+uuid.NewString()+"/hide", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, fiber.StatusNotFound)
	}

	var body response.Response
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Message != "issue not found" {
		t.Fatalf("unexpected message: %q", body.Message)
	}
}

func TestAdminHandlerBanDeviceReturns404ForMissingDevice(t *testing.T) {
	app := fiber.New()
	handler := NewAdminHandler(&adminServiceHTTPFake{banErr: service.ErrModerationTargetNotFound}, nil, false)

	app.Post("/admin/devices/:id/ban", func(c *fiber.Ctx) error {
		c.Locals("admin_username", "admin")
		return handler.BanDevice(c)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/admin/devices/"+uuid.NewString()+"/ban", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, fiber.StatusNotFound)
	}

	var body response.Response
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Message != "device not found" {
		t.Fatalf("unexpected message: %q", body.Message)
	}
}

func TestAdminHandlerLoginSetsSessionCookie(t *testing.T) {
	app := fiber.New()
	handler := NewAdminHandler(&adminServiceHTTPFake{loginToken: "session-token"}, nil, false)

	app.Post("/admin/login", handler.Login)

	req := httptest.NewRequest(
		fiber.MethodPost,
		"/admin/login",
		bytes.NewBufferString(`{"username":"moderator","password":"super-secret-123"}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, fiber.StatusOK)
	}

	setCookie := resp.Header.Get("Set-Cookie")
	if !strings.Contains(setCookie, service.AdminSessionCookieName+"=session-token") {
		t.Fatalf("expected session cookie to be set, got %q", setCookie)
	}
	if !strings.Contains(strings.ToLower(setCookie), "httponly") {
		t.Fatalf("expected session cookie to be HttpOnly, got %q", setCookie)
	}
	if !strings.Contains(strings.ToLower(setCookie), "samesite=strict") {
		t.Fatalf("expected SameSite=Strict cookie, got %q", setCookie)
	}
}

func TestAdminHandlerLogoutRevokesServerSessionAndClearsCookie(t *testing.T) {
	app := fiber.New()
	serviceFake := &adminServiceHTTPFake{}
	handler := NewAdminHandler(serviceFake, nil, false)

	app.Post("/admin/logout", func(c *fiber.Ctx) error {
		c.Locals("admin_session_token", "session-token")
		return handler.Logout(c)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/admin/logout", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, fiber.StatusOK)
	}
	if serviceFake.revokedToken != "session-token" {
		t.Fatalf("expected session token to be revoked, got %q", serviceFake.revokedToken)
	}

	setCookie := strings.ToLower(resp.Header.Get("Set-Cookie"))
	if !strings.Contains(setCookie, service.AdminSessionCookieName+"=") {
		t.Fatalf("expected session cookie clearing header, got %q", setCookie)
	}
	if !strings.Contains(setCookie, "max-age=0") && !strings.Contains(setCookie, "expires=thu, 01 jan 1970") {
		t.Fatalf("expected cleared cookie, got %q", setCookie)
	}
}

type adminServiceHTTPFake struct {
	loginToken   string
	loginErr     error
	hideErr      error
	banErr       error
	revokedToken string
}

func (f *adminServiceHTTPFake) Login(_, _, _ string) (string, error) { return f.loginToken, f.loginErr }

func (f *adminServiceHTTPFake) ValidateSession(_ string) *service.AdminSession { return nil }

func (f *adminServiceHTTPFake) RevokeSession(token string) { f.revokedToken = token }

func (f *adminServiceHTTPFake) ListIssues(_ context.Context, _ int, _ int, _ *string) ([]*domain.AdminIssue, error) {
	return nil, nil
}

func (f *adminServiceHTTPFake) GetIssueDetail(_ context.Context, _ uuid.UUID) (*domain.AdminIssueDetail, error) {
	return nil, nil
}

func (f *adminServiceHTTPFake) HideIssue(_ context.Context, _ uuid.UUID, _ string, _ *string) error {
	return f.hideErr
}

func (f *adminServiceHTTPFake) UnhideIssue(_ context.Context, _ uuid.UUID, _ string, _ *string) error {
	return nil
}

func (f *adminServiceHTTPFake) FixIssue(_ context.Context, _ uuid.UUID, _ string, _ *string) error {
	return nil
}

func (f *adminServiceHTTPFake) RejectIssue(_ context.Context, _ uuid.UUID, _ string, _ *string) error {
	return nil
}

func (f *adminServiceHTTPFake) BanDevice(_ context.Context, _ uuid.UUID, _ string, _ *string) error {
	return f.banErr
}
