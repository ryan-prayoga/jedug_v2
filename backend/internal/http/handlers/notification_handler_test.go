package handlers

import (
	"io"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"jedug_backend/internal/domain"
)

func TestParseNotificationReplayCursorReadsQuery(t *testing.T) {
	app := fiber.New()
	app.Get("/notifications/stream", func(c *fiber.Ctx) error {
		cursor, err := parseNotificationReplayCursor(c)
		if err != nil {
			return err
		}
		return c.SendString("cursor=" + c.Query("last_event_id") + "|" + strconv.FormatInt(cursor, 10))
	})

	req := httptest.NewRequest(fiber.MethodGet, "/notifications/stream?last_event_id=42", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	body := string(bodyBytes)
	if !strings.Contains(body, "42|42") {
		t.Fatalf("unexpected response body: %q", body)
	}
}

func TestParseNotificationReplayCursorRejectsInvalidValue(t *testing.T) {
	app := fiber.New()
	app.Get("/notifications/stream", func(c *fiber.Ctx) error {
		_, err := parseNotificationReplayCursor(c)
		if err == nil {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	})

	req := httptest.NewRequest(fiber.MethodGet, "/notifications/stream?last_event_id=oops", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestNotificationReplayMessageIncludesSSEEventID(t *testing.T) {
	item := &domain.Notification{
		ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		IssueID:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		EventID:   99,
		Type:      "status_updated",
		Title:     "Status berubah",
		Message:   "Ada perubahan status",
		CreatedAt: time.Unix(1_700_000_000, 0).UTC(),
	}

	msg, err := notificationReplayMessage(item)
	if err != nil {
		t.Fatalf("notificationReplayMessage error: %v", err)
	}
	if !strings.Contains(msg, "id: 99\n") {
		t.Fatalf("expected SSE id frame, got %q", msg)
	}
	if !strings.Contains(msg, "event: notification\n") {
		t.Fatalf("expected notification event frame, got %q", msg)
	}
}
