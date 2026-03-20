package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
	"jedug_backend/internal/sse"
)

type NotificationHandler struct {
	svc     service.NotificationService
	authSvc service.FollowerAuthService
}

func NewNotificationHandler(svc service.NotificationService, authSvc service.FollowerAuthService) *NotificationHandler {
	return &NotificationHandler{svc: svc, authSvc: authSvc}
}

// GET /api/v1/notifications?follower_token=<token>&limit=50
func (h *NotificationHandler) List(c *fiber.Ctx) error {
	followerID, err := authenticateFollowerToken(c, h.authSvc)
	if err != nil {
		return mapFollowerAuthError(c, err)
	}

	limit := c.QueryInt("limit", 50)
	notifications, err := h.svc.GetByFollowerID(c.Context(), followerID, limit)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch notifications")
	}

	return response.OK(c, fiber.Map{
		"items": notifications,
	})
}

// PATCH /api/v1/notifications/:id/read?follower_token=<token>
func (h *NotificationHandler) MarkRead(c *fiber.Ctx) error {
	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "notification id must be a valid UUID")
	}

	followerID, err := authenticateFollowerToken(c, h.authSvc)
	if err != nil {
		return mapFollowerAuthError(c, err)
	}

	readAt, updated, err := h.svc.MarkAsRead(c.Context(), notificationID, followerID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to mark notification as read")
	}
	if !updated {
		return response.Error(c, fiber.StatusNotFound, "notification not found for this follower")
	}

	return response.OK(c, fiber.Map{
		"read_at": readAt,
	})
}

// DELETE /api/v1/notifications/:id?follower_token=<token>
func (h *NotificationHandler) Delete(c *fiber.Ctx) error {
	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "notification id must be a valid UUID")
	}

	followerID, err := authenticateFollowerToken(c, h.authSvc)
	if err != nil {
		return mapFollowerAuthError(c, err)
	}

	deleted, err := h.svc.Delete(c.Context(), notificationID, followerID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to delete notification")
	}

	return response.OK(c, fiber.Map{
		"deleted": deleted,
	})
}

// GET /api/v1/notifications/stream?stream_token=<token>
// Opens a Server-Sent Events stream.  The connection is kept alive via 30-second
// ping events.  When the client disconnects, the cleanup function removes the
// subscription from the in-process SSE hub.
func (h *NotificationHandler) Stream(c *fiber.Ctx) error {
	followerID, err := authenticateFollowerStreamToken(c, h.authSvc)
	if err != nil {
		return mapFollowerAuthError(c, err)
	}

	lastEventID, err := parseNotificationReplayCursor(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "last_event_id must be a valid integer")
	}

	replayItems, err := h.svc.GetByFollowerIDSinceEventID(c.Context(), followerID, lastEventID, 50)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to prepare notification replay")
	}

	followerKey := followerID.String()
	startedAt := time.Now().UTC()
	log.Printf(
		"[SSE] stream_open follower=%s rid=%s replayed=%d last_event_id=%d active_connections=%d dropped_total=%d",
		followerKey,
		requestID(c),
		len(replayItems),
		lastEventID,
		sse.Default.ConnectionCount(),
		sse.Default.DroppedCount(),
	)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	// Tell nginx not to buffer the SSE stream (see nginx config guidance in docs).
	c.Set("X-Accel-Buffering", "no")

	ch, done := sse.Default.Subscribe(followerKey)
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		closeReason := "client_disconnect"
		defer func() {
			done()
			log.Printf(
				"[SSE] stream_close follower=%s duration=%s reason=%s active_connections=%d dropped_total=%d",
				followerKey,
				time.Since(startedAt).Round(time.Second),
				closeReason,
				sse.Default.ConnectionCount(),
				sse.Default.DroppedCount(),
			)
		}()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		if _, err := fmt.Fprintf(w, "retry: 1000\n"); err != nil {
			closeReason = "initial_write_failed"
			return
		}
		if _, err := fmt.Fprintf(w, "event: connected\ndata: {\"replayed\":%d}\n\n", len(replayItems)); err != nil {
			closeReason = "connected_event_failed"
			return
		}
		if err := w.Flush(); err != nil {
			closeReason = "initial_flush_failed"
			return
		}

		for _, item := range replayItems {
			msg, msgErr := notificationReplayMessage(item)
			if msgErr != nil {
				continue
			}
			if _, err := fmt.Fprint(w, msg); err != nil {
				closeReason = "replay_write_failed"
				return
			}
			if err := w.Flush(); err != nil {
				closeReason = "replay_flush_failed"
				return
			}
		}

		for {
			select {
			case msg := <-ch:
				if _, err := fmt.Fprint(w, msg); err != nil {
					closeReason = "stream_write_failed"
					return
				}
				if err := w.Flush(); err != nil {
					closeReason = "stream_flush_failed"
					return
				}
			case <-ticker.C:
				if _, err := fmt.Fprintf(w, "event: ping\ndata: {}\n\n"); err != nil {
					closeReason = "ping_write_failed"
					return
				}
				if err := w.Flush(); err != nil {
					closeReason = "ping_flush_failed"
					return
				}
			}
		}
	}))
	return nil
}

type notificationStreamPayload struct {
	ID        string    `json:"id"`
	IssueID   string    `json:"issue_id"`
	EventID   int64     `json:"event_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func parseNotificationReplayCursor(c *fiber.Ctx) (int64, error) {
	raw := c.Query("last_event_id")
	if raw == "" {
		raw = c.Get("Last-Event-ID")
	}
	if raw == "" {
		return 0, nil
	}

	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("invalid last_event_id")
	}
	return parsed, nil
}

func notificationReplayMessage(item *domain.Notification) (string, error) {
	payload, err := json.Marshal(notificationStreamPayload{
		ID:        item.ID.String(),
		IssueID:   item.IssueID.String(),
		EventID:   item.EventID,
		Type:      item.Type,
		Title:     item.Title,
		Message:   item.Message,
		CreatedAt: item.CreatedAt,
	})
	if err != nil {
		return "", err
	}

	return sse.FormatEvent("notification", payload, strconv.FormatInt(item.EventID, 10)), nil
}
