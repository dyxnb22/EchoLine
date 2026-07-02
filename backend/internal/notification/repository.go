package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Event is a notification event record.
type Event struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      string
	Payload   json.RawMessage
	ReadAt    *time.Time
	CreatedAt time.Time
}

// Repository manages notification events.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a notification repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a notification event.
func (r *Repository) Create(ctx context.Context, userID uuid.UUID, typ string, payload any) (*Event, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	const q = `
		INSERT INTO notification_events (id, user_id, type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, type, payload, read_at, created_at
	`
	now := time.Now().UTC()
	id := uuid.New()
	row := r.pool.QueryRow(ctx, q, id, userID, typ, raw, now)
	var e Event
	if err := row.Scan(&e.ID, &e.UserID, &e.Type, &e.Payload, &e.ReadAt, &e.CreatedAt); err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return &e, nil
}

// ListForUser returns the most recent notifications for a user (limit 50).
func (r *Repository) ListForUser(ctx context.Context, userID uuid.UUID) ([]Event, error) {
	const q = `
		SELECT id, user_id, type, payload, read_at, created_at
		FROM notification_events
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.UserID, &e.Type, &e.Payload, &e.ReadAt, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// MarkRead marks a notification as read.
func (r *Repository) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	const q = `
		UPDATE notification_events
		SET read_at = $1
		WHERE id = $2 AND user_id = $3 AND read_at IS NULL
	`
	_, err := r.pool.Exec(ctx, q, time.Now().UTC(), id, userID)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	return nil
}

// MarkAllRead marks all unread notifications as read for a user.
func (r *Repository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	const q = `
		UPDATE notification_events
		SET read_at = $1
		WHERE user_id = $2 AND read_at IS NULL
	`
	_, err := r.pool.Exec(ctx, q, time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}
