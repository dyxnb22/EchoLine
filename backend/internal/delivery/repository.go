package delivery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidTransition = errors.New("invalid delivery status transition")

// Status is delivery/read state.
type Status string

const (
	StatusSent      Status = "sent"
	StatusDelivered Status = "delivered"
	StatusRead      Status = "read"
)

var statusRank = map[Status]int{
	StatusSent:      1,
	StatusDelivered: 2,
	StatusRead:      3,
}

// Record is a per-user/device delivery state.
type Record struct {
	MessageID uuid.UUID
	UserID    uuid.UUID
	DeviceID  string
	Status    Status
	AckedAt   time.Time
}

// Repository stores delivery acknowledgements.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a delivery repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// UpsertACK records delivery state, only allowing forward transitions.
func (r *Repository) UpsertACK(ctx context.Context, messageID, userID uuid.UUID, deviceID string, status Status) (*Record, error) {
	if deviceID == "" {
		deviceID = "default"
	}

	existing, err := r.Get(ctx, messageID, userID, deviceID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	now := time.Now().UTC()
	if existing != nil {
		if statusRank[status] < statusRank[existing.Status] {
			return existing, ErrInvalidTransition
		}
		if statusRank[status] == statusRank[existing.Status] {
			return existing, nil
		}
	}

	const q = `
		INSERT INTO message_deliveries (message_id, user_id, device_id, status, acked_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (message_id, user_id, device_id)
		DO UPDATE SET status = EXCLUDED.status, acked_at = EXCLUDED.acked_at
		WHERE message_deliveries.status IS DISTINCT FROM EXCLUDED.status
		   AND (
		     CASE message_deliveries.status
		       WHEN 'sent' THEN 1
		       WHEN 'delivered' THEN 2
		       WHEN 'read' THEN 3
		     END
		   ) <= (
		     CASE EXCLUDED.status
		       WHEN 'sent' THEN 1
		       WHEN 'delivered' THEN 2
		       WHEN 'read' THEN 3
		     END
		   )
		RETURNING message_id, user_id, device_id, status, acked_at
	`
	var rec Record
	var statusStr string
	err = r.pool.QueryRow(ctx, q, messageID, userID, deviceID, string(status), now).Scan(
		&rec.MessageID, &rec.UserID, &rec.DeviceID, &statusStr, &rec.AckedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) && existing != nil {
			return existing, nil
		}
		return nil, fmt.Errorf("upsert ack: %w", err)
	}
	rec.Status = Status(statusStr)
	return &rec, nil
}

// Get loads a delivery record.
func (r *Repository) Get(ctx context.Context, messageID, userID uuid.UUID, deviceID string) (*Record, error) {
	if deviceID == "" {
		deviceID = "default"
	}
	const q = `
		SELECT message_id, user_id, device_id, status, acked_at
		FROM message_deliveries
		WHERE message_id = $1 AND user_id = $2 AND device_id = $3
	`
	var rec Record
	var statusStr string
	err := r.pool.QueryRow(ctx, q, messageID, userID, deviceID).Scan(
		&rec.MessageID, &rec.UserID, &rec.DeviceID, &statusStr, &rec.AckedAt,
	)
	if err != nil {
		return nil, err
	}
	rec.Status = Status(statusStr)
	return &rec, nil
}
