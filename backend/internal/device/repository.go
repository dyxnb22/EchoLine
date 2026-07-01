package device

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists devices.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a device repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Upsert creates or updates a device record for a user.
func (r *Repository) Upsert(ctx context.Context, userID uuid.UUID, deviceName, platform string) (*Device, error) {
	now := time.Now().UTC()
	id := uuid.New()

	const q = `
		INSERT INTO devices (id, user_id, device_name, platform, last_seen_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id, user_id, device_name, platform, last_seen_at, created_at
	`

	row := r.pool.QueryRow(ctx, q, id, userID, deviceName, platform, now)
	var d Device
	if err := row.Scan(&d.ID, &d.UserID, &d.DeviceName, &d.Platform, &d.LastSeenAt, &d.CreatedAt); err != nil {
		return nil, fmt.Errorf("upsert device: %w", err)
	}
	return &d, nil
}

// GetByID loads a device by ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Device, error) {
	const q = `
		SELECT id, user_id, device_name, platform, last_seen_at, created_at
		FROM devices
		WHERE id = $1
	`
	var d Device
	err := r.pool.QueryRow(ctx, q, id).Scan(&d.ID, &d.UserID, &d.DeviceName, &d.Platform, &d.LastSeenAt, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}
	return &d, nil
}
