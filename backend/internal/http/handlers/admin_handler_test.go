package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

func TestAdminHandlerHideIssueReturns404ForMissingIssue(t *testing.T) {
	app := fiber.New()
	handler := NewAdminHandler(&adminServiceHTTPFake{hideErr: service.ErrModerationTargetNotFound}, nil)

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
	handler := NewAdminHandler(&adminServiceHTTPFake{banErr: service.ErrModerationTargetNotFound}, nil)

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

type adminServiceHTTPFake struct {
	hideErr error
	banErr  error
}

func (f *adminServiceHTTPFake) Login(_, _ string) (string, error) { return "", nil }

func (f *adminServiceHTTPFake) ValidateSession(_ string) *service.AdminSession { return nil }

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
