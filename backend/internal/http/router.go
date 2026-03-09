package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/config"
	"jedug_backend/internal/http/handlers"
	"jedug_backend/internal/http/middleware"
	"jedug_backend/internal/repository"
	"jedug_backend/internal/service"
	"jedug_backend/internal/storage"
)

func NewRouter(cfg *config.Config, db *pgxpool.Pool) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:   cfg.AppName,
		BodyLimit: 25 * 1024 * 1024, // 25 MB to accommodate file uploads
	})

	middleware.Register(app, cfg.CORSAllowOrigins)

	// Storage driver (local for now; swap in R2Driver here when ready)
	store := storage.NewLocalDriver(cfg.StoragePublicBaseURL, cfg.UploadDir)

	// Serve uploaded files statically
	app.Static("/uploads/gallery", cfg.UploadDir)

	// Wire up dependencies
	deviceRepo := repository.NewDeviceRepository(db)
	issueRepo := repository.NewIssueRepository(db)
	reportRepo := repository.NewReportRepository(db)
	adminRepo := repository.NewAdminRepository(db)

	deviceSvc := service.NewDeviceService(deviceRepo)
	issueSvc := service.NewIssueService(issueRepo)
	reportSvc := service.NewReportService(deviceRepo, reportRepo)
	adminSvc := service.NewAdminService(cfg.AdminUsername, cfg.AdminPassword, adminRepo)

	healthHandler := handlers.NewHealthHandler(db)
	deviceHandler := handlers.NewDeviceHandler(deviceSvc)
	issueHandler := handlers.NewIssueHandler(issueSvc, store)
	uploadHandler := handlers.NewUploadHandler(store)
	reportHandler := handlers.NewReportHandler(reportSvc)
	adminHandler := handlers.NewAdminHandler(adminSvc, store)

	// Routes
	api := app.Group("/api/v1")

	api.Get("/health", healthHandler.Health)

	device := api.Group("/device")
	device.Post("/bootstrap", deviceHandler.Bootstrap)
	device.Post("/consent", deviceHandler.Consent)

	uploads := api.Group("/uploads")
	uploads.Post("/presign", uploadHandler.Presign)
	uploads.Post("/file/*", uploadHandler.UploadFile)

	api.Post("/reports", reportHandler.Submit)

	issues := api.Group("/issues")
	issues.Get("/", issueHandler.List)
	issues.Get("/:id", issueHandler.Get)

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

	return app
}

