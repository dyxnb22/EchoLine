package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CursorRepository persists per-device sync cursors.
type CursorRepository struct {
	pool *pgxpool.Pool
}

// NewCursorRepository creates a cursor repository.
func NewCursorRepository(pool *pgxpool.Pool) *CursorRepository {
	return &CursorRepository{pool: pool}
}

// Upsert stores the latest seq seen by a device in a conversation.
func (r *CursorRepository) Upsert(ctx context.Context, userID uuid.UUID, deviceID string, conversationID uuid.UUID, lastSeq int64) error {
	const q = `
		INSERT INTO device_sync_cursors (user_id, device_id, conversation_id, last_seq, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, device_id, conversation_id)
		DO UPDATE SET last_seq = GREATEST(device_sync_cursors.last_seq, EXCLUDED.last_seq), updated_at = EXCLUDED.updated_at
	`
	_, err := r.pool.Exec(ctx, q, userID, deviceID, conversationID, lastSeq, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("upsert device cursor: %w", err)
	}
	return nil
}

// ListForDevice returns stored cursors for a user/device pair.
func (r *CursorRepository) ListForDevice(ctx context.Context, userID uuid.UUID, deviceID string) (map[uuid.UUID]int64, error) {
	const q = `
		SELECT conversation_id, last_seq
		FROM device_sync_cursors
		WHERE user_id = $1 AND device_id = $2
	`
	rows, err := r.pool.Query(ctx, q, userID, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[uuid.UUID]int64)
	for rows.Next() {
		var convID uuid.UUID
		var seq int64
		if err := rows.Scan(&convID, &seq); err != nil {
			return nil, err
		}
		out[convID] = seq
	}
	return out, rows.Err()
}
