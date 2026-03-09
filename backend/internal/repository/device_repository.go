package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
)

type DeviceRepository interface {
	FindByTokenHash(ctx context.Context, tokenHash string) (*domain.Device, error)
	Create(ctx context.Context, device *domain.Device) error
	UpdateLastSeen(ctx context.Context, id uuid.UUID) error
	CreateConsent(ctx context.Context, consent *domain.DeviceConsent) error
}

type deviceRepository struct {
	db *pgxpool.Pool
}

func NewDeviceRepository(db *pgxpool.Pool) DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.Device, error) {
	query := `
		SELECT id, anon_token_hash, trust_score, is_banned,
		       last_ip::text, last_user_agent,
		       first_seen_at, last_seen_at, created_at, updated_at
		FROM devices
		WHERE anon_token_hash = $1
		LIMIT 1
	`
	var d domain.Device
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&d.ID, &d.AnonTokenHash, &d.TrustScore, &d.IsBanned,
		&d.LastIP, &d.LastUserAgent,
		&d.FirstSeenAt, &d.LastSeenAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *deviceRepository) Create(ctx context.Context, device *domain.Device) error {
	// first_seen_at, last_seen_at, created_at, updated_at all default to NOW()
	query := `
		INSERT INTO devices (id, anon_token_hash, last_ip, last_user_agent)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query,
		device.ID, device.AnonTokenHash, device.LastIP, device.LastUserAgent,
	)
	return err
}

func (r *deviceRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE devices SET last_seen_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *deviceRepository) CreateConsent(ctx context.Context, consent *domain.DeviceConsent) error {
	// id is BIGSERIAL, omitted from INSERT
	query := `
		INSERT INTO device_consents (device_id, terms_version, privacy_version, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		consent.DeviceID, consent.TermsVersion, consent.PrivacyVersion,
		consent.IPAddress, consent.UserAgent,
	)
	return err
}
