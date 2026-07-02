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

// TouchByClientDevice upserts last_seen for a user/device client id string.
func (r *Repository) TouchByClientDevice(ctx context.Context, userID uuid.UUID, clientDeviceID, platform string) error {
	if platform == "" {
		platform = "web"
	}
	now := time.Now().UTC()
	const updateQ = `
		UPDATE devices
		SET last_seen_at = $3, platform = $4
		WHERE user_id = $1 AND device_name = $2
	`
	tag, err := r.pool.Exec(ctx, updateQ, userID, clientDeviceID, now, platform)
	if err != nil {
		return fmt.Errorf("touch device: %w", err)
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	const insertQ = `
		INSERT INTO devices (id, user_id, device_name, platform, last_seen_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`
	_, err = r.pool.Exec(ctx, insertQ, uuid.New(), userID, clientDeviceID, platform, now)
	return err
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
