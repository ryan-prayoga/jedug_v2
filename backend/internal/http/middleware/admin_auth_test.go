package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
)

type adminAuthServiceStub struct {
	session *service.AdminSession
}

func (f *adminAuthServiceStub) Login(_, _, _ string) (string, error) { return "", nil }

func (f *adminAuthServiceStub) ValidateSession(token string) *service.AdminSession {
	if token == "session-token" {
		return f.session
	}
	return nil
}

func (f *adminAuthServiceStub) RevokeSession(_ string) {}

func (f *adminAuthServiceStub) ListIssues(_ context.Context, _ int, _ int, _ *string) ([]*domain.AdminIssue, error) {
	return nil, nil
}

func (f *adminAuthServiceStub) GetIssueDetail(_ context.Context, _ uuid.UUID) (*domain.AdminIssueDetail, error) {
	return nil, nil
}

func (f *adminAuthServiceStub) HideIssue(_ context.Context, _ uuid.UUID, _ string, _ *string) error {
	return nil
}

func (f *adminAuthServiceStub) UnhideIssue(_ context.Context, _ uuid.UUID, _ string, _ *string) error {
	return nil
}

func (f *adminAuthServiceStub) FixIssue(_ context.Context, _ uuid.UUID, _ string, _ *string) error {
	return nil
}

func (f *adminAuthServiceStub) RejectIssue(_ context.Context, _ uuid.UUID, _ string, _ *string) error {
	return nil
}

func (f *adminAuthServiceStub) BanDevice(_ context.Context, _ uuid.UUID, _ string, _ *string) error {
	return nil
}

func TestAdminAuthReadsSessionCookie(t *testing.T) {
	app := fiber.New()
	app.Get("/admin/me", AdminAuth(&adminAuthServiceStub{
		session: &service.AdminSession{Username: "moderator"},
	}), func(c *fiber.Ctx) error {
		return response.OK(c, fiber.Map{"username": c.Locals("admin_username")})
	})

	req := httptest.NewRequest(fiber.MethodGet, "/admin/me", nil)
	req.Header.Set("Cookie", service.AdminSessionCookieName+"=session-token")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, fiber.StatusOK)
	}
}
