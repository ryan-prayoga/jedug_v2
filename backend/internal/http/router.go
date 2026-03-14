package http

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/config"
	"jedug_backend/internal/http/handlers"
	"jedug_backend/internal/http/middleware"
	"jedug_backend/internal/repository"
	"jedug_backend/internal/service"
	"jedug_backend/internal/storage"
)

func NewRouter(cfg *config.Config, db *pgxpool.Pool) (*fiber.App, error) {
	app := fiber.New(fiber.Config{
		AppName:   cfg.AppName,
		BodyLimit: 25 * 1024 * 1024, // 25 MB to accommodate file uploads
	})

	middleware.Register(app, cfg.CORSAllowOrigins)

	legacyLocal := storage.NewLocalDriver(cfg.StoragePublicBaseURL, cfg.UploadDir)

	var activeStorage storage.Driver
	switch cfg.StorageDriver {
	case "", "local":
		activeStorage = legacyLocal
	case "r2":
		r2Driver, err := storage.NewR2Driver(context.Background(), storage.R2Config{
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

	store := storage.NewService(activeStorage, legacyLocal)

	// Keep serving legacy local uploads even after STORAGE_DRIVER switches to r2.
	app.Static("/uploads/gallery", cfg.UploadDir)

	// Wire up dependencies
	deviceRepo := repository.NewDeviceRepository(db)
	issueRepo := repository.NewIssueRepository(db)
	issueFollowRepo := repository.NewIssueFollowRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	reportRepo := repository.NewReportRepository(db, repository.ReportRepositoryConfig{
		DuplicateRadiusM: cfg.DuplicateRadiusM,
	})
	adminRepo := repository.NewAdminRepository(db)
	flagRepo := repository.NewFlagRepository(db)
	locationRepo := repository.NewLocationRepository(db)

	deviceSvc := service.NewDeviceService(deviceRepo)
	issueSvc := service.NewIssueService(issueRepo)
	issueFollowSvc := service.NewIssueFollowService(issueRepo, issueFollowRepo)
	statsSvc := service.NewStatsService(statsRepo)
	reverseGeocoder := service.NewHTTPReverseGeocoder(
		cfg.ReverseGeocodeEnabled,
		cfg.ReverseGeocodeURL,
		cfg.ReverseGeocodeUserAgent,
		cfg.ReverseGeocodeTimeout,
		cfg.ReverseGeocodeCacheTTL,
	)
	locationNormalizer := service.NewReportLocationNormalizer(locationRepo, reverseGeocoder)
	reportSvc := service.NewReportService(deviceRepo, reportRepo, locationNormalizer)
	adminSvc := service.NewAdminService(cfg.AdminUsername, cfg.AdminPassword, adminRepo)
	flagSvc := service.NewFlagService(deviceRepo, flagRepo, adminRepo)
	locationSvc := service.NewLocationService(locationRepo, reverseGeocoder)

	healthHandler := handlers.NewHealthHandler(db)
	deviceHandler := handlers.NewDeviceHandler(deviceSvc)
	issueHandler := handlers.NewIssueHandler(issueSvc, store)
	issueFollowHandler := handlers.NewIssueFollowHandler(issueFollowSvc)
	statsHandler := handlers.NewStatsHandler(statsSvc)
	uploadHandler := handlers.NewUploadHandler(store)
	reportHandler := handlers.NewReportHandler(reportSvc)
	adminHandler := handlers.NewAdminHandler(adminSvc, store)
	flagHandler := handlers.NewFlagHandler(flagSvc)
	locationHandler := handlers.NewLocationHandler(locationSvc)

	// Rate limiters
	rlBootstrap := middleware.RateLimit(10, 1*time.Minute)
	rlConsent := middleware.RateLimit(10, 1*time.Minute)
	rlPresign := middleware.RateLimit(20, 1*time.Minute)
	rlReport := middleware.RateLimit(5, 1*time.Minute)
	rlFlag := middleware.RateLimit(10, 1*time.Minute)
	rlFollow := middleware.RateLimit(30, 1*time.Minute)

	// Routes
	api := app.Group("/api/v1")

	api.Get("/health", healthHandler.Health)

	device := api.Group("/device")
	device.Post("/bootstrap", rlBootstrap, deviceHandler.Bootstrap)
	device.Post("/consent", rlConsent, deviceHandler.Consent)

	uploads := api.Group("/uploads")
	uploads.Post("/presign", rlPresign, uploadHandler.Presign)
	uploads.Post("/file/*", uploadHandler.UploadFile)

	api.Post("/reports", rlReport, reportHandler.Submit)

	location := api.Group("/location")
	location.Get("/label", locationHandler.ResolveLabel)

	issues := api.Group("/issues")
	issues.Get("/", issueHandler.List)
	issues.Get("/:id/timeline", issueHandler.Timeline)
	issues.Post("/:id/follow", rlFollow, issueFollowHandler.Follow)
	issues.Delete("/:id/follow", rlFollow, issueFollowHandler.Unfollow)
	issues.Post("/:id/followers", rlFollow, issueFollowHandler.Follow)
	issues.Delete("/:id/followers", rlFollow, issueFollowHandler.Unfollow)
	issues.Get("/:id/followers/count", issueFollowHandler.Count)
	issues.Get("/:id/count", issueFollowHandler.Count)
	issues.Get("/:id/follow-status", issueFollowHandler.Status)
	issues.Get("/:id/follow/status", issueFollowHandler.Status)
	issues.Get("/:id", issueHandler.Get)
	issues.Post("/:id/flag", rlFlag, flagHandler.FlagIssue)

	api.Get("/stats", statsHandler.Get)

	// Admin routes
	admin := api.Group("/admin")
	admin.Post("/login", adminHandler.Login)

	// Protected admin routes
	adminAuth := admin.Group("", middleware.AdminAuth(adminSvc))
	adminAuth.Get("/me", adminHandler.Me)
	adminAuth.Get("/issues", adminHandler.ListIssues)
	adminAuth.Get("/issues/:id", adminHandler.GetIssue)
	adminAuth.Post("/issues/:id/hide", adminHandler.HideIssue)
	adminAuth.Post("/issues/:id/fix", adminHandler.FixIssue)
	adminAuth.Post("/issues/:id/reject", adminHandler.RejectIssue)
	adminAuth.Post("/issues/:id/unhide", adminHandler.UnhideIssue)
	adminAuth.Post("/devices/:id/ban", adminHandler.BanDevice)

	return app, nil
}
