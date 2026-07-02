package pin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAlreadyPinned = errors.New("message already pinned")
var ErrNotPinned = errors.New("message is not pinned")

// PinnedMessage is a pinned message record.
type PinnedMessage struct {
	ConversationID uuid.UUID
	MessageID      uuid.UUID
	PinnedBy       uuid.UUID
	PinnedAt       time.Time
}

// Repository manages pinned messages.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a pin repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Pin pins a message in a conversation.
func (r *Repository) Pin(ctx context.Context, convID, messageID, pinnedBy uuid.UUID) (*PinnedMessage, error) {
	const q = `
		INSERT INTO pinned_messages (conversation_id, message_id, pinned_by, pinned_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (conversation_id, message_id) DO NOTHING
		RETURNING conversation_id, message_id, pinned_by, pinned_at
	`
	now := time.Now().UTC()
	row := r.pool.QueryRow(ctx, q, convID, messageID, pinnedBy, now)
	var p PinnedMessage
	if err := row.Scan(&p.ConversationID, &p.MessageID, &p.PinnedBy, &p.PinnedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAlreadyPinned
		}
		return nil, fmt.Errorf("pin message: %w", err)
	}
	return &p, nil
}

// Unpin removes a pinned message.
func (r *Repository) Unpin(ctx context.Context, convID, messageID uuid.UUID) error {
	const q = `
		DELETE FROM pinned_messages
		WHERE conversation_id = $1 AND message_id = $2
	`
	tag, err := r.pool.Exec(ctx, q, convID, messageID)
	if err != nil {
		return fmt.Errorf("unpin message: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotPinned
	}
	return nil
}

// List returns all pinned messages for a conversation.
func (r *Repository) List(ctx context.Context, convID uuid.UUID) ([]PinnedMessage, error) {
	const q = `
		SELECT conversation_id, message_id, pinned_by, pinned_at
		FROM pinned_messages
		WHERE conversation_id = $1
		ORDER BY pinned_at DESC
	`
	rows, err := r.pool.Query(ctx, q, convID)
	if err != nil {
		return nil, fmt.Errorf("list pinned: %w", err)
	}
	defer rows.Close()

	var pins []PinnedMessage
	for rows.Next() {
		var p PinnedMessage
		if err := rows.Scan(&p.ConversationID, &p.MessageID, &p.PinnedBy, &p.PinnedAt); err != nil {
			return nil, fmt.Errorf("scan pinned: %w", err)
		}
		pins = append(pins, p)
	}
	return pins, rows.Err()
}
