package repository

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LocationLabel struct {
	RegionID        int64
	RegionName      string
	RegionLevel     string
	ParentName      *string
	GrandparentName *string
}

type LocationRepository interface {
	ResolveLabelByPoint(ctx context.Context, longitude, latitude float64) (*LocationLabel, error)
}

type locationRepository struct {
	db *pgxpool.Pool
}

func NewLocationRepository(db *pgxpool.Pool) LocationRepository {
	return &locationRepository{db: db}
}

// ResolveLabelByPoint picks the smallest polygon region that covers the point.
// This allows UX label fallback even when exact district-level data is unavailable.
func (r *locationRepository) ResolveLabelByPoint(ctx context.Context, longitude, latitude float64) (*LocationLabel, error) {
	query := `
		WITH input_point AS (
			SELECT ST_SetSRID(ST_MakePoint($1, $2), 4326) AS geom
		), best_region AS (
			SELECT reg.id, reg.name, reg.level, reg.parent_id
			FROM regions reg
			CROSS JOIN input_point p
			WHERE ST_Covers(reg.geom, p.geom)
			ORDER BY ST_Area(reg.geom::geography) ASC
			LIMIT 1
		)
		SELECT
			b.id,
			b.name,
			b.level,
			parent.name AS parent_name,
			grandparent.name AS grandparent_name
		FROM best_region b
		LEFT JOIN regions parent ON parent.id = b.parent_id
		LEFT JOIN regions grandparent ON grandparent.id = parent.parent_id
	`

	var out LocationLabel
	err := r.db.QueryRow(ctx, query, longitude, latitude).Scan(
		&out.RegionID,
		&out.RegionName,
		&out.RegionLevel,
		&out.ParentName,
		&out.GrandparentName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		log.Printf("[LOCATION_LABEL] internal_query_error lon=%.6f lat=%.6f err=%v", longitude, latitude, err)
		return nil, err
	}

	return &out, nil
}
