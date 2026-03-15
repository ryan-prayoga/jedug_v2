package handlers

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/service"
	"jedug_backend/internal/sse"
)

type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// GET /api/v1/notifications?follower_id=<uuid>&limit=50
func (h *NotificationHandler) List(c *fiber.Ctx) error {
	followerID, err := uuid.Parse(c.Query("follower_id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "follower_id must be a valid UUID")
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

// PATCH /api/v1/notifications/:id/read?follower_id=<uuid>
func (h *NotificationHandler) MarkRead(c *fiber.Ctx) error {
	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "notification id must be a valid UUID")
	}

	followerID, err := uuid.Parse(c.Query("follower_id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "follower_id must be a valid UUID")
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

// DELETE /api/v1/notifications/:id?follower_id=<uuid>
func (h *NotificationHandler) Delete(c *fiber.Ctx) error {
	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "notification id must be a valid UUID")
	}

	followerID, err := uuid.Parse(c.Query("follower_id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "follower_id must be a valid UUID")
	}

	deleted, err := h.svc.Delete(c.Context(), notificationID, followerID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to delete notification")
	}

	return response.OK(c, fiber.Map{
		"deleted": deleted,
	})
}

// GET /api/v1/notifications/stream?follower_id=<uuid>
// Opens a Server-Sent Events stream.  The connection is kept alive via 30-second
// ping events.  When the client disconnects, the cleanup function removes the
// subscription from the in-process SSE hub.
func (h *NotificationHandler) Stream(c *fiber.Ctx) error {
	followerID, err := uuid.Parse(c.Query("follower_id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "follower_id must be a valid UUID")
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	// Tell nginx not to buffer the SSE stream (see nginx config guidance in docs).
	c.Set("X-Accel-Buffering", "no")

	ch, done := sse.Default.Subscribe(followerID.String())
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		defer done()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		if _, err := fmt.Fprintf(w, "event: connected\ndata: {}\n\n"); err != nil {
			return
		}
		if err := w.Flush(); err != nil {
			return
		}

		for {
			select {
			case msg := <-ch:
				if _, err := fmt.Fprint(w, msg); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			case <-ticker.C:
				if _, err := fmt.Fprintf(w, "event: ping\ndata: {}\n\n"); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	}))
	return nil
}
