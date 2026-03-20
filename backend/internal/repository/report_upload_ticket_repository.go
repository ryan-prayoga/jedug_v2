package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateReportUploadTicketInput struct {
	ObjectKey   string
	DeviceID    uuid.UUID
	ContentType string
	SizeBytes   int
	UploadMode  string
	ExpiresAt   time.Time
}

type ReportUploadTicket struct {
	ObjectKey   string
	DeviceID    uuid.UUID
	ContentType string
	SizeBytes   int
	UploadMode  string
	IssuedAt    time.Time
	ExpiresAt   time.Time
}

type OrphanedReportUpload struct {
	ObjectKey  string
	UploadMode string
	IssuedAt   time.Time
}

type ReportUploadTicketRepository interface {
	CreateOrReplace(ctx context.Context, input CreateReportUploadTicketInput) error
	FindByObjectKey(ctx context.Context, objectKey string) (*ReportUploadTicket, error)
	CountPendingByDeviceSince(ctx context.Context, deviceID uuid.UUID, since time.Time) (int, error)
	ListOrphansBefore(ctx context.Context, cutoff time.Time, limit int) ([]OrphanedReportUpload, error)
	DeleteByObjectKeys(ctx context.Context, objectKeys []string) (int64, error)
	CountOrphansBefore(ctx context.Context, cutoff time.Time) (int64, error)
}

type reportUploadTicketRepository struct {
	db *pgxpool.Pool
}

func NewReportUploadTicketRepository(db *pgxpool.Pool) ReportUploadTicketRepository {
	return &reportUploadTicketRepository{db: db}
}

func (r *reportUploadTicketRepository) CreateOrReplace(ctx context.Context, input CreateReportUploadTicketInput) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO report_upload_tickets (
			object_key, device_id, content_type, size_bytes, upload_mode, issued_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, NOW(), $6)
		ON CONFLICT (object_key) DO UPDATE
		SET device_id = EXCLUDED.device_id,
		    content_type = EXCLUDED.content_type,
		    size_bytes = EXCLUDED.size_bytes,
		    upload_mode = EXCLUDED.upload_mode,
		    issued_at = NOW(),
		    expires_at = EXCLUDED.expires_at
	`, input.ObjectKey, input.DeviceID, input.ContentType, input.SizeBytes, input.UploadMode, input.ExpiresAt.UTC())
	return err
}

func (r *reportUploadTicketRepository) FindByObjectKey(ctx context.Context, objectKey string) (*ReportUploadTicket, error) {
	var ticket ReportUploadTicket
	err := r.db.QueryRow(ctx, `
		SELECT object_key, device_id, content_type, size_bytes, upload_mode, issued_at, expires_at
		FROM report_upload_tickets
		WHERE object_key = $1
	`, objectKey).Scan(
		&ticket.ObjectKey,
		&ticket.DeviceID,
		&ticket.ContentType,
		&ticket.SizeBytes,
		&ticket.UploadMode,
		&ticket.IssuedAt,
		&ticket.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ticket, nil
}

func (r *reportUploadTicketRepository) CountPendingByDeviceSince(ctx context.Context, deviceID uuid.UUID, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM report_upload_tickets
		WHERE device_id = $1
		  AND issued_at >= $2
	`, deviceID, since.UTC()).Scan(&count)
	return count, err
}

func (r *reportUploadTicketRepository) ListOrphansBefore(ctx context.Context, cutoff time.Time, limit int) ([]OrphanedReportUpload, error) {
	rows, err := r.db.Query(ctx, `
		SELECT object_key, upload_mode, issued_at
		FROM report_upload_tickets
		WHERE issued_at < $1
		ORDER BY issued_at ASC
		LIMIT $2
	`, cutoff.UTC(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orphaned := make([]OrphanedReportUpload, 0, limit)
	for rows.Next() {
		var item OrphanedReportUpload
		if err := rows.Scan(&item.ObjectKey, &item.UploadMode, &item.IssuedAt); err != nil {
			return nil, err
		}
		orphaned = append(orphaned, item)
	}
	return orphaned, rows.Err()
}

func (r *reportUploadTicketRepository) DeleteByObjectKeys(ctx context.Context, objectKeys []string) (int64, error) {
	result, err := r.db.Exec(ctx, `
		DELETE FROM report_upload_tickets
		WHERE object_key = ANY($1::text[])
	`, objectKeys)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *reportUploadTicketRepository) CountOrphansBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM report_upload_tickets
		WHERE issued_at < $1
	`, cutoff.UTC()).Scan(&count)
	return count, err
}
