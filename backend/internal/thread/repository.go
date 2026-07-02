package thread

import (
	"context"
	"fmt"

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

// ListReplies returns direct replies to a parent message within a conversation.
func (r *Repository) ListReplies(ctx context.Context, convID, parentMsgID uuid.UUID) ([]message.Message, error) {
	const q = `
		SELECT id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
		FROM messages
		WHERE parent_message_id = $1 AND conversation_id = $2
		ORDER BY seq ASC
	`
	rows, err := r.pool.Query(ctx, q, parentMsgID, convID)
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
