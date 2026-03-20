package ops

import (
	"context"
	"log"
	"time"

	"jedug_backend/internal/storage"
)

type Runner struct {
	store    *Store
	storage  *storage.Service
	policy   RetentionPolicy
	interval time.Duration
	enabled  bool
}

func NewRunner(store *Store, storageSvc *storage.Service, policy RetentionPolicy, interval time.Duration, enabled bool) *Runner {
	return &Runner{
		store:    store,
		storage:  storageSvc,
		policy:   policy,
		interval: interval,
		enabled:  enabled,
	}
}

func (r *Runner) Start(ctx context.Context) {
	if !r.enabled {
		log.Printf("[OPS] retention_runner_disabled")
		return
	}

	go func() {
		r.runCycle("startup")

		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.runCycle("interval")
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (r *Runner) RunOnce(ctx context.Context) (*CleanupSummary, error) {
	if !r.enabled {
		return &CleanupSummary{Skipped: true}, nil
	}

	summary, err := r.store.RunCleanup(ctx, r.policy)
	if err != nil || summary == nil || summary.Skipped {
		return summary, err
	}

	orphansDeleted, err := r.cleanupUploadOrphans(ctx)
	if err != nil {
		return nil, err
	}
	summary.UploadOrphansDeleted = orphansDeleted
	return summary, nil
}

func (r *Runner) runCycle(trigger string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	summary, err := r.RunOnce(ctx)
	if err != nil {
		log.Printf("[OPS] retention_failed trigger=%s err=%v", trigger, err)
		return
	}
	if summary == nil || summary.Skipped {
		log.Printf("[OPS] retention_skipped trigger=%s reason=lock_not_acquired", trigger)
		return
	}

	log.Printf(
		"[OPS] retention_completed trigger=%s notifications_deleted=%d push_subscriptions_disabled=%d push_subscriptions_deleted=%d push_deliveries_delivered_deleted=%d push_deliveries_failed_deleted=%d upload_orphans_deleted=%d",
		trigger,
		summary.NotificationsDeleted,
		summary.PushSubscriptionsDisabled,
		summary.PushSubscriptionsDeleted,
		summary.PushDeliveriesDeliveredDeleted,
		summary.PushDeliveriesFailedDeleted,
		summary.UploadOrphansDeleted,
	)
}

func (r *Runner) cleanupUploadOrphans(ctx context.Context) (int64, error) {
	if r.storage == nil || r.policy.UploadOrphanRetention <= 0 {
		return 0, nil
	}

	const batchSize = 100
	cutoff := time.Now().UTC().Add(-r.policy.UploadOrphanRetention)
	var totalDeleted int64

	for {
		items, err := r.store.ListUploadOrphansBefore(ctx, cutoff, batchSize)
		if err != nil {
			return totalDeleted, err
		}
		if len(items) == 0 {
			return totalDeleted, nil
		}

		deletableKeys := make([]string, 0, len(items))
		for _, item := range items {
			if err := r.storage.Delete(ctx, item.UploadMode, item.ObjectKey); err != nil {
				log.Printf("[OPS] upload_orphan_delete_failed object_key=%s upload_mode=%s err=%v", item.ObjectKey, item.UploadMode, err)
				continue
			}
			deletableKeys = append(deletableKeys, item.ObjectKey)
		}

		deleted, err := r.store.DeleteUploadOrphans(ctx, deletableKeys)
		if err != nil {
			return totalDeleted, err
		}
		totalDeleted += deleted

		if len(items) < batchSize {
			return totalDeleted, nil
		}
	}
}
