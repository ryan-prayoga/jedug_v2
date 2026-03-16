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
	DistrictName    *string
	RegencyName     *string
	ProvinceName    *string
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
		), resolved AS (
			SELECT
				b.id,
				b.name,
				CASE
					WHEN LOWER(COALESCE(b.level, '')) IN ('province', 'provinsi') THEN 'province'
					WHEN LOWER(COALESCE(b.level, '')) IN ('city', 'kota') THEN 'city'
					WHEN LOWER(COALESCE(b.level, '')) IN ('regency', 'kabupaten') THEN 'regency'
					WHEN LOWER(COALESCE(b.level, '')) IN ('district', 'kecamatan') THEN 'district'
					WHEN LOWER(COALESCE(b.level, '')) IN ('subdistrict') THEN 'subdistrict'
					WHEN LOWER(COALESCE(b.level, '')) IN ('village', 'kelurahan', 'desa') THEN 'village'
					ELSE LOWER(COALESCE(b.level, ''))
				END AS level,
				parent.name AS parent_name,
				grandparent.name AS grandparent_name,
				CASE
					WHEN LOWER(COALESCE(b.level, '')) IN ('district', 'kecamatan', 'subdistrict') THEN b.name
					WHEN LOWER(COALESCE(parent.level, '')) IN ('district', 'kecamatan', 'subdistrict') THEN parent.name
					WHEN LOWER(COALESCE(grandparent.level, '')) IN ('district', 'kecamatan', 'subdistrict') THEN grandparent.name
					WHEN LOWER(COALESCE(great_grandparent.level, '')) IN ('district', 'kecamatan', 'subdistrict') THEN great_grandparent.name
					WHEN LOWER(COALESCE(great_great_grandparent.level, '')) IN ('district', 'kecamatan', 'subdistrict') THEN great_great_grandparent.name
					ELSE NULL
				END AS district_name,
				CASE
					WHEN LOWER(COALESCE(b.level, '')) IN ('city', 'kota', 'regency', 'kabupaten') THEN b.name
					WHEN LOWER(COALESCE(parent.level, '')) IN ('city', 'kota', 'regency', 'kabupaten') THEN parent.name
					WHEN LOWER(COALESCE(grandparent.level, '')) IN ('city', 'kota', 'regency', 'kabupaten') THEN grandparent.name
					WHEN LOWER(COALESCE(great_grandparent.level, '')) IN ('city', 'kota', 'regency', 'kabupaten') THEN great_grandparent.name
					WHEN LOWER(COALESCE(great_great_grandparent.level, '')) IN ('city', 'kota', 'regency', 'kabupaten') THEN great_great_grandparent.name
					ELSE NULL
				END AS regency_name,
				CASE
					WHEN LOWER(COALESCE(b.level, '')) IN ('province', 'provinsi') THEN b.name
					WHEN LOWER(COALESCE(parent.level, '')) IN ('province', 'provinsi') THEN parent.name
					WHEN LOWER(COALESCE(grandparent.level, '')) IN ('province', 'provinsi') THEN grandparent.name
					WHEN LOWER(COALESCE(great_grandparent.level, '')) IN ('province', 'provinsi') THEN great_grandparent.name
					WHEN LOWER(COALESCE(great_great_grandparent.level, '')) IN ('province', 'provinsi') THEN great_great_grandparent.name
					ELSE NULL
				END AS province_name
			FROM best_region b
			LEFT JOIN regions parent ON parent.id = b.parent_id
			LEFT JOIN regions grandparent ON grandparent.id = parent.parent_id
			LEFT JOIN regions great_grandparent ON great_grandparent.id = grandparent.parent_id
			LEFT JOIN regions great_great_grandparent ON great_great_grandparent.id = great_grandparent.parent_id
		)
		SELECT
			resolved.id,
			resolved.name,
			resolved.level,
			resolved.parent_name,
			resolved.grandparent_name,
			resolved.district_name,
			resolved.regency_name,
			resolved.province_name
		FROM resolved
	`

	var out LocationLabel
	err := r.db.QueryRow(ctx, query, longitude, latitude).Scan(
		&out.RegionID,
		&out.RegionName,
		&out.RegionLevel,
		&out.ParentName,
		&out.GrandparentName,
		&out.DistrictName,
		&out.RegencyName,
		&out.ProvinceName,
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
