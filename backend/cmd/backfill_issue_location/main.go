package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"jedug_backend/internal/config"
	"jedug_backend/internal/database"
	"jedug_backend/internal/repository"
	"jedug_backend/internal/service"
)

type issueLocationRow struct {
	ID        uuid.UUID
	Longitude float64
	Latitude  float64
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	var (
		limit  = flag.Int("limit", 200, "maximum number of issues to backfill")
		dryRun = flag.Bool("dry-run", false, "preview updates without writing to database")
	)
	flag.Parse()

	cfg := config.Load()
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()

	reverseGeocoder := service.NewHTTPReverseGeocoder(
		cfg.ReverseGeocodeEnabled,
		cfg.ReverseGeocodeURL,
		cfg.ReverseGeocodeUserAgent,
		cfg.ReverseGeocodeTimeout,
		cfg.ReverseGeocodeCacheTTL,
	)
	locationRepo := repository.NewLocationRepository(db)
	normalizer := service.NewReportLocationNormalizer(locationRepo, reverseGeocoder)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rows, err := db.Query(ctx, `
		SELECT
			i.id,
			ST_X(i.public_location::geometry) AS longitude,
			ST_Y(i.public_location::geometry) AS latitude
		FROM issues i
		WHERE (i.road_name IS NULL OR BTRIM(i.road_name) = '')
		   OR i.region_id IS NULL
		ORDER BY i.created_at ASC
		LIMIT $1
	`, *limit)
	if err != nil {
		log.Fatalf("query issues to backfill: %v", err)
	}
	defer rows.Close()

	items := make([]issueLocationRow, 0, *limit)
	for rows.Next() {
		var row issueLocationRow
		if err := rows.Scan(&row.ID, &row.Longitude, &row.Latitude); err != nil {
			log.Fatalf("scan issue row: %v", err)
		}
		items = append(items, row)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("iterate issue rows: %v", err)
	}

	if len(items) == 0 {
		log.Println("No issues need location backfill")
		return
	}

	log.Printf("Found %d issues to process (dry-run=%t)", len(items), *dryRun)

	updated := 0
	for _, item := range items {
		norm := normalizer.NormalizeForReport(ctx, item.Longitude, item.Latitude)
		roadName := norm.RoadName
		regionID := norm.RegionID

		if *dryRun {
			log.Printf(
				"[dry-run] issue=%s road=%q region_id=%v",
				item.ID.String(),
				stringOrEmpty(roadName),
				int64OrNil(regionID),
			)
			continue
		}

		tag, execErr := db.Exec(ctx, `
			UPDATE issues
			SET
				road_name = CASE
					WHEN (road_name IS NULL OR BTRIM(road_name) = '')
					     AND $2 IS NOT NULL AND BTRIM($2) <> ''
					THEN $2
					ELSE road_name
				END,
				region_id = CASE
					WHEN region_id IS NULL AND $3 IS NOT NULL
					THEN $3
					ELSE region_id
				END,
				updated_at = NOW()
			WHERE id = $1
			  AND (
					((road_name IS NULL OR BTRIM(road_name) = '') AND $2 IS NOT NULL AND BTRIM($2) <> '')
					OR (region_id IS NULL AND $3 IS NOT NULL)
			  )
		`, item.ID, roadName, regionID)
		if execErr != nil {
			log.Printf("update failed issue=%s err=%v", item.ID, execErr)
			continue
		}

		if tag.RowsAffected() > 0 {
			updated++
		}
	}

	if *dryRun {
		log.Printf("Dry-run completed for %d issues", len(items))
		return
	}

	log.Printf("Backfill completed: updated %d/%d issues", updated, len(items))
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func int64OrNil(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}
