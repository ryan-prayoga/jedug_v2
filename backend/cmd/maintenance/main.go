package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"jedug_backend/internal/config"
	"jedug_backend/internal/database"
	"jedug_backend/internal/ops"
	"jedug_backend/internal/storage"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	cfg := config.Load()
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	storeService, err := buildStorageService(context.Background(), cfg)
	if err != nil {
		log.Fatalf("failed to initialize storage service: %v", err)
	}

	store := ops.NewStore(db)
	runner := ops.NewRunner(store, storeService, ops.RetentionPolicy{
		NotificationsRetention:             cfg.NotificationsRetention,
		PushSubscriptionsStaleAfter:        cfg.PushSubscriptionsStaleAfter,
		PushSubscriptionsDisabledRetention: cfg.PushSubscriptionsDisabledRetention,
		PushDeliveryDeliveredRetention:     cfg.PushDeliveryDeliveredRetention,
		PushDeliveryFailedRetention:        cfg.PushDeliveryFailedRetention,
		UploadOrphanRetention:              cfg.UploadOrphanRetention,
	}, cfg.MaintenanceInterval, cfg.MaintenanceEnabled)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	summary, err := runner.RunOnce(ctx)
	if err != nil {
		log.Fatalf("maintenance failed: %v", err)
	}
	if summary == nil || summary.Skipped {
		log.Printf("[OPS] maintenance skipped")
		return
	}

	log.Printf(
		"[OPS] maintenance completed notifications_deleted=%d push_subscriptions_disabled=%d push_subscriptions_deleted=%d push_deliveries_delivered_deleted=%d push_deliveries_failed_deleted=%d upload_orphans_deleted=%d",
		summary.NotificationsDeleted,
		summary.PushSubscriptionsDisabled,
		summary.PushSubscriptionsDeleted,
		summary.PushDeliveriesDeliveredDeleted,
		summary.PushDeliveriesFailedDeleted,
		summary.UploadOrphansDeleted,
	)
}

func buildStorageService(ctx context.Context, cfg *config.Config) (*storage.Service, error) {
	legacyLocal := storage.NewLocalDriver(cfg.StoragePublicBaseURL, cfg.UploadDir)

	var activeStorage storage.Driver
	switch cfg.StorageDriver {
	case "", "local":
		activeStorage = legacyLocal
	case "r2":
		r2Driver, err := storage.NewR2Driver(ctx, storage.R2Config{
			AccountID:       cfg.R2AccountID,
			AccessKeyID:     cfg.R2AccessKeyID,
			SecretAccessKey: cfg.R2SecretAccessKey,
			Bucket:          cfg.R2Bucket,
			Endpoint:        cfg.R2Endpoint,
			PublicBaseURL:   cfg.R2PublicBaseURL,
		})
		if err != nil {
			return nil, fmt.Errorf("init r2 storage: %w", err)
		}
		activeStorage = r2Driver
	default:
		return nil, fmt.Errorf("unsupported STORAGE_DRIVER: %s", cfg.StorageDriver)
	}

	return storage.NewService(activeStorage, legacyLocal), nil
}
