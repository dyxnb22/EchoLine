package thread

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/message"
)

// Repository persists threaded replies.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a thread repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// SendReply persists a message with parent_message_id set.
func (r *Repository) SendReply(ctx context.Context, convID, senderID, parentMsgID uuid.UUID, body string) (*message.Message, error) {
	const q = `
		INSERT INTO messages (id, conversation_id, sender_id, client_msg_id, seq, type, body, status, parent_message_id, created_at, updated_at)
		VALUES (
			gen_random_uuid(),
			$1, $2,
			'',
			(SELECT COALESCE(MAX(seq), 0) + 1 FROM messages WHERE conversation_id = $1),
			'text',
			$3,
			'sent',
			$4,
			$5, $5
		)
		RETURNING id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
	`
	now := time.Now().UTC()
	row := r.pool.QueryRow(ctx, q, convID, senderID, body, parentMsgID, now)
	var msg message.Message
	if err := row.Scan(
		&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.ClientMsgID,
		&msg.Seq, &msg.Type, &msg.Body, &msg.Status,
		&msg.CreatedAt, &msg.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("send reply: %w", err)
	}
	return &msg, nil
}

// ListReplies returns all direct replies to a parent message.
func (r *Repository) ListReplies(ctx context.Context, parentMsgID uuid.UUID) ([]message.Message, error) {
	const q = `
		SELECT id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
		FROM messages
		WHERE parent_message_id = $1
		ORDER BY seq ASC
	`
	rows, err := r.pool.Query(ctx, q, parentMsgID)
	if err != nil {
		return nil, fmt.Errorf("list replies: %w", err)
	}
	defer rows.Close()

	var out []message.Message
	for rows.Next() {
		var msg message.Message
		if err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.ClientMsgID,
			&msg.Seq, &msg.Type, &msg.Body, &msg.Status,
			&msg.CreatedAt, &msg.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reply: %w", err)
		}
		out = append(out, msg)
	}
	return out, rows.Err()
}
