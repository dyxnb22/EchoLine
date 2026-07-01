package reaction

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Reaction is a single emoji reaction on a message.
type Reaction struct {
	MessageID uuid.UUID
	UserID    uuid.UUID
	Emoji     string
	CreatedAt time.Time
}

// Repository persists message reactions.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a reaction repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Add inserts or ignores a reaction.
func (r *Repository) Add(ctx context.Context, messageID, userID uuid.UUID, emoji string) (*Reaction, error) {
	const q = `
		INSERT INTO message_reactions (message_id, user_id, emoji, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (message_id, user_id, emoji) DO NOTHING
		RETURNING message_id, user_id, emoji, created_at
	`
	now := time.Now().UTC()
	row := r.pool.QueryRow(ctx, q, messageID, userID, emoji, now)
	var rx Reaction
	if err := row.Scan(&rx.MessageID, &rx.UserID, &rx.Emoji, &rx.CreatedAt); err != nil {
		return nil, fmt.Errorf("add reaction: %w", err)
	}
	return &rx, nil
}

// Remove deletes a reaction.
func (r *Repository) Remove(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	const q = `DELETE FROM message_reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3`
	_, err := r.pool.Exec(ctx, q, messageID, userID, emoji)
	if err != nil {
		return fmt.Errorf("remove reaction: %w", err)
	}
	return nil
}

// ListByMessage returns all reactions for a message.
func (r *Repository) ListByMessage(ctx context.Context, messageID uuid.UUID) ([]Reaction, error) {
	const q = `
		SELECT message_id, user_id, emoji, created_at
		FROM message_reactions
		WHERE message_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, q, messageID)
	if err != nil {
		return nil, fmt.Errorf("list reactions: %w", err)
	}
	defer rows.Close()

	var out []Reaction
	for rows.Next() {
		var rx Reaction
		if err := rows.Scan(&rx.MessageID, &rx.UserID, &rx.Emoji, &rx.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan reaction: %w", err)
		}
		out = append(out, rx)
	}
	return out, rows.Err()
}
