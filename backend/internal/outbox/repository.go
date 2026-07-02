package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Event is a pending outbox row.
type Event struct {
	ID        uuid.UUID
	Topic     string
	Payload   []byte
	Attempts  int
	CreatedAt time.Time
}

// Repository stores transactional outbox events.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates an outbox repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// EnqueueInTx inserts an outbox event in the caller transaction.
func (r *Repository) EnqueueInTx(ctx context.Context, tx execer, topic string, payload []byte) error {
	const q = `
		INSERT INTO outbox_events (id, topic, payload, status, attempts, created_at)
		VALUES ($1, $2, $3, 'pending', 0, $4)
	`
	_, err := tx.Exec(ctx, q, uuid.New(), topic, payload, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("enqueue outbox: %w", err)
	}
	return nil
}

// FetchPending claims pending events by moving them to processing in one transaction.
func (r *Repository) FetchPending(ctx context.Context, limit int) ([]Event, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin outbox tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
		UPDATE outbox_events
		SET status = 'processing', processing_at = $2
		WHERE id IN (
			SELECT id
			FROM outbox_events
			WHERE status = 'pending'
			ORDER BY created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, topic, payload::text, attempts, created_at
	`
	now := time.Now().UTC()
	rows, err := tx.Query(ctx, q, limit, now)
	if err != nil {
		return nil, fmt.Errorf("claim pending outbox: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var payload string
		if err := rows.Scan(&e.ID, &e.Topic, &payload, &e.Attempts, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan outbox: %w", err)
		}
		e.Payload = []byte(payload)
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return events, nil
}

// CountPending returns pending outbox rows for metrics.
func (r *Repository) CountPending(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(*) FROM outbox_events WHERE status = 'pending'`
	var count int64
	err := r.pool.QueryRow(ctx, q).Scan(&count)
	return count, err
}

// Enqueue inserts an outbox event outside a caller transaction.
func (r *Repository) Enqueue(ctx context.Context, topic string, payload []byte) error {
	const q = `
		INSERT INTO outbox_events (id, topic, payload, status, attempts, created_at)
		VALUES ($1, $2, $3, 'pending', 0, $4)
	`
	_, err := r.pool.Exec(ctx, q, uuid.New(), topic, payload, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("enqueue outbox: %w", err)
	}
	return nil
}

// MarkPublished marks an event as published.
func (r *Repository) MarkPublished(ctx context.Context, id uuid.UUID) error {
	const q = `
		UPDATE outbox_events
		SET status = 'published', published_at = $2
		WHERE id = $1 AND status IN ('pending', 'processing')
	`
	_, err := r.pool.Exec(ctx, q, id, time.Now().UTC())
	return err
}

// CleanupPublished deletes published outbox rows older than the cutoff.
func (r *Repository) CleanupPublished(ctx context.Context, olderThan time.Time) (int64, error) {
	const q = `
		DELETE FROM outbox_events
		WHERE status = 'published' AND published_at IS NOT NULL AND published_at < $1
	`
	tag, err := r.pool.Exec(ctx, q, olderThan)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// RequeueStaleProcessing resets stuck processing rows older than the cutoff.
func (r *Repository) RequeueStaleProcessing(ctx context.Context, olderThan time.Time) (int64, error) {
	const q = `
		UPDATE outbox_events
		SET status = 'pending', processing_at = NULL
		WHERE status = 'processing' AND processing_at IS NOT NULL AND processing_at < $1
	`
	tag, err := r.pool.Exec(ctx, q, olderThan)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// MarkFailed increments attempts, moves to DLQ after threshold, or leaves pending.
func (r *Repository) MarkFailed(ctx context.Context, id uuid.UUID) error {
	const fetchQ = `
		SELECT topic, payload::text, attempts
		FROM outbox_events
		WHERE id = $1 AND status IN ('pending', 'processing')
	`
	var topic, payload string
	var attempts int
	if err := r.pool.QueryRow(ctx, fetchQ, id).Scan(&topic, &payload, &attempts); err != nil {
		return err
	}

	nextAttempts := attempts + 1
	if nextAttempts >= 5 {
		const dlqQ = `
			INSERT INTO dead_letter_events (id, source_topic, payload, error_message, attempts, created_at)
			VALUES ($1, $2, $3::jsonb, 'publish failed', $4, $5)
		`
		now := time.Now().UTC()
		if _, err := r.pool.Exec(ctx, dlqQ, uuid.New(), topic, payload, nextAttempts, now); err != nil {
			return err
		}
		const failQ = `UPDATE outbox_events SET attempts = $2, status = 'failed' WHERE id = $1`
		_, err := r.pool.Exec(ctx, failQ, id, nextAttempts)
		return err
	}

	const q = `UPDATE outbox_events SET attempts = attempts + 1, status = 'pending', processing_at = NULL WHERE id = $1 AND status IN ('pending', 'processing')`
	_, err := r.pool.Exec(ctx, q, id)
	return err
}
