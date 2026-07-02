package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Hit is a search result row.
type Hit struct {
	MessageID      uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Body           string
	Seq            int64
	CreatedAt      time.Time
}

// Repository indexes and queries messages.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a search repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// IndexMessage upserts a message into the search index.
func (r *Repository) IndexMessage(ctx context.Context, messageID, conversationID, senderID uuid.UUID, body string, seq int64, createdAt time.Time) error {
	const q = `
		INSERT INTO message_search_index (message_id, conversation_id, sender_id, body, seq, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (message_id) DO UPDATE
		SET body = EXCLUDED.body, seq = EXCLUDED.seq
	`
	_, err := r.pool.Exec(ctx, q, messageID, conversationID, senderID, body, seq, createdAt)
	if err != nil {
		return fmt.Errorf("index message: %w", err)
	}
	return nil
}

// Search finds messages in conversations the user is a member of.
func (r *Repository) Search(ctx context.Context, userID uuid.UUID, query string, limit int) ([]Hit, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	const q = `
		SELECT m.message_id, m.conversation_id, m.sender_id, m.body, m.seq, m.created_at
		FROM message_search_index m
		JOIN conversation_members cm ON cm.conversation_id = m.conversation_id AND cm.user_id = $1
		WHERE m.search_vector @@ plainto_tsquery('simple', $2)
		ORDER BY m.created_at DESC
		LIMIT $3
	`
	rows, err := r.pool.Query(ctx, q, userID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	defer rows.Close()

	var hits []Hit
	for rows.Next() {
		var h Hit
		if err := rows.Scan(&h.MessageID, &h.ConversationID, &h.SenderID, &h.Body, &h.Seq, &h.CreatedAt); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	return hits, rows.Err()
}

// UpdateMessageBody updates indexed text for an edited message.
func (r *Repository) UpdateMessageBody(ctx context.Context, messageID uuid.UUID, body string) error {
	const q = `UPDATE message_search_index SET body = $2 WHERE message_id = $1`
	_, err := r.pool.Exec(ctx, q, messageID, body)
	return err
}

// DeleteMessage removes a message from the search index.
func (r *Repository) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	const q = `DELETE FROM message_search_index WHERE message_id = $1`
	_, err := r.pool.Exec(ctx, q, messageID)
	return err
}
