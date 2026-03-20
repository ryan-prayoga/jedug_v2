package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/http/response"
	"jedug_backend/internal/ops"
	"jedug_backend/internal/sse"
)

type HealthHandler struct {
	db        *pgxpool.Pool
	opsStore  *ops.Store
	sseHub    *sse.Hub
	policy    ops.RetentionPolicy
	startedAt time.Time
}

func NewHealthHandler(db *pgxpool.Pool, opsStore *ops.Store, sseHub *sse.Hub, policy ops.RetentionPolicy) *HealthHandler {
	return &HealthHandler{
		db:        db,
		opsStore:  opsStore,
		sseHub:    sseHub,
		policy:    policy,
		startedAt: time.Now().UTC(),
	}
}

func (h *HealthHandler) Health(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "database unreachable")
	}

	snapshot, err := h.opsStore.HealthSnapshot(ctx, h.policy)
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "health snapshot unavailable")
	}

	return response.OK(c, fiber.Map{
		"status":     "ok",
		"checked_at": time.Now().UTC(),
		"uptime_sec": int(time.Since(h.startedAt).Seconds()),
		"checks": fiber.Map{
			"database": "ok",
		},
		"runtime": fiber.Map{
			"sse_followers":        h.sseHub.FollowerCount(),
			"sse_connections":      h.sseHub.ConnectionCount(),
			"sse_dropped_total":    h.sseHub.DroppedCount(),
			"push_ready":           snapshot.PushReadyCount,
			"push_failed_last_24h": snapshot.PushFailedLast24H,
		},
		"retention": fiber.Map{
			"notifications_over_retention":  snapshot.NotificationsOverRetention,
			"stale_push_subscriptions":      snapshot.StalePushSubscriptions,
			"disabled_push_subscriptions":   snapshot.DisabledPushSubscriptions,
			"upload_orphans_over_retention": snapshot.UploadOrphansOverRetention,
			"issue_events_policy":           "keep",
		},
		"tables": fiber.Map{
			"issue_events_estimate":          snapshot.IssueEventsEstimate,
			"notifications_estimate":         snapshot.NotificationsEstimate,
			"push_subscriptions_estimate":    snapshot.PushSubscriptionsEstimate,
			"push_delivery_jobs_estimate":    snapshot.PushDeliveryJobsEstimate,
			"report_upload_tickets_estimate": snapshot.ReportUploadTicketsEstimate,
		},
	})
}
