package main

import (
	"context"
	"log"
	"time"

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
		log.Fatalf("preflight database connect failed: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("preflight database ping failed: %v", err)
	}

	if _, err := apphttp.NewRouter(cfg, db); err != nil {
		log.Fatalf("preflight router initialization failed: %v", err)
	}

	log.Println("preflight ok")
}
