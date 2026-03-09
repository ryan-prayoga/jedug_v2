package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"jedug_backend/internal/config"
	"jedug_backend/internal/database"
	apphttp "jedug_backend/internal/http"
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

	app, err := apphttp.NewRouter(cfg, db)
	if err != nil {
		log.Fatalf("failed to build router: %v", err)
	}

	go func() {
		if err := app.Listen(":" + cfg.AppPort); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	_ = app.Shutdown()
}
