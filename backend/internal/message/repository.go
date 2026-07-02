package message

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/eventbus"
	"github.com/echoline/echoline/backend/internal/outbox"
)

// Repository persists messages.
type Repository struct {
	pool   *pgxpool.Pool
	outbox *outbox.Repository
}

// NewRepository creates a message repository.
func NewRepository(pool *pgxpool.Pool, outboxRepo *outbox.Repository) *Repository {
	return &Repository{pool: pool, outbox: outboxRepo}
}

// Create inserts a message and allocates the next conversation seq in one transaction.
func (r *Repository) Create(ctx context.Context, conversationID, senderID uuid.UUID, clientMsgID string, msgType Type, body string, attachmentID *uuid.UUID) (*Message, error) {
	clientMsgID = strings.TrimSpace(clientMsgID)
	body = strings.TrimSpace(body)
	if body == "" && attachmentID == nil {
		return nil, fmt.Errorf("message body is required")
	}
	if msgType == "" {
		msgType = TypeText
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var nextSeq int64
	const seqQ = `
		UPDATE conversations
		SET latest_seq = latest_seq + 1, updated_at = $2
		WHERE id = $1
		RETURNING latest_seq
	`
	now := time.Now().UTC()
	if err := tx.QueryRow(ctx, seqQ, conversationID, now).Scan(&nextSeq); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("conversation not found")
		}
		return nil, fmt.Errorf("allocate seq: %w", err)
	}

	msgID := uuid.New()
	const insertQ = `
		INSERT INTO messages (
			id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'normal', $8, $9)
		RETURNING id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
	`

	row := tx.QueryRow(ctx, insertQ, msgID, conversationID, senderID, clientMsgID, nextSeq, string(msgType), body, now, now)
	msg, err := scanMessage(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if clientMsgID != "" {
				existing, lookupErr := r.getByClientMsgID(ctx, tx, senderID, clientMsgID)
				if lookupErr == nil {
					if err := tx.Commit(ctx); err == nil {
						return existing, nil
					}
				}
			}
			return nil, ErrDuplicateClientID
		}
		return nil, fmt.Errorf("insert message: %w", err)
	}

	const updateConv = `
		UPDATE conversations
		SET last_message_id = $2
		WHERE id = $1
	`
	if _, err := tx.Exec(ctx, updateConv, conversationID, msg.ID); err != nil {
		return nil, fmt.Errorf("update conversation last message: %w", err)
	}

	if attachmentID != nil {
		const linkQ = `
			UPDATE attachments
			SET message_id = $2
			WHERE id = $1 AND message_id IS NULL
		`
		tag, err := tx.Exec(ctx, linkQ, *attachmentID, msg.ID)
		if err != nil {
			return nil, fmt.Errorf("link attachment: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return nil, fmt.Errorf("attachment not available")
		}
	}

	if r.outbox != nil {
		payloadMap := map[string]any{
			"id":              msg.ID.String(),
			"conversation_id": msg.ConversationID.String(),
			"sender_id":       msg.SenderID.String(),
			"seq":             msg.Seq,
			"type":            msg.Type,
			"body":            msg.Body,
			"created_at":      msg.CreatedAt.UTC().Format(time.RFC3339),
		}
		if attachmentID != nil {
			payloadMap["attachment_id"] = attachmentID.String()
		}
		payload, err := json.Marshal(payloadMap)
		if err != nil {
			return nil, fmt.Errorf("marshal outbox payload: %w", err)
		}
		if err := r.outbox.EnqueueInTx(ctx, tx, eventbus.TopicMessageCreated, payload); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit message: %w", err)
	}

	return msg, nil
}

// List returns messages for a conversation ordered by seq descending.
func (r *Repository) List(ctx context.Context, conversationID uuid.UUID, beforeSeq *int64, limit int) ([]Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var rows pgx.Rows
	var err error
	if beforeSeq != nil {
		const q = `
			SELECT id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
			FROM messages
			WHERE conversation_id = $1 AND seq < $2
			ORDER BY seq DESC
			LIMIT $3
		`
		rows, err = r.pool.Query(ctx, q, conversationID, *beforeSeq, limit)
	} else {
		const q = `
			SELECT id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
			FROM messages
			WHERE conversation_id = $1
			ORDER BY seq DESC
			LIMIT $2
		`
		rows, err = r.pool.Query(ctx, q, conversationID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		msg, err := scanMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, *msg)
	}
	return messages, rows.Err()
}

// ListSince returns messages with seq strictly greater than afterSeq.
func (r *Repository) ListSince(ctx context.Context, conversationID uuid.UUID, afterSeq int64, limit int) ([]Message, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	const q = `
		SELECT id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
		FROM messages
		WHERE conversation_id = $1 AND seq > $2
		ORDER BY seq ASC
		LIMIT $3
	`
	rows, err := r.pool.Query(ctx, q, conversationID, afterSeq, limit)
	if err != nil {
		return nil, fmt.Errorf("list since seq: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		msg, err := scanMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, *msg)
	}
	return messages, rows.Err()
}

// GetByID returns a message by id within a conversation.
func (r *Repository) GetByID(ctx context.Context, conversationID, messageID uuid.UUID) (*Message, error) {
	const q = `
		SELECT id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
		FROM messages
		WHERE id = $1 AND conversation_id = $2
	`
	row := r.pool.QueryRow(ctx, q, messageID, conversationID)
	msg, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return msg, nil
}

// Edit updates message body for the sender.
func (r *Repository) Edit(ctx context.Context, conversationID, messageID, senderID uuid.UUID, body string) (*Message, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, fmt.Errorf("message body is required")
	}
	const q = `
		UPDATE messages
		SET body = $4, status = 'edited', updated_at = $5
		WHERE id = $1 AND conversation_id = $2 AND sender_id = $3 AND status = 'normal'
		RETURNING id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
	`
	now := time.Now().UTC()
	row := r.pool.QueryRow(ctx, q, messageID, conversationID, senderID, body, now)
	msg, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return msg, nil
}

// Recall marks a message as recalled.
func (r *Repository) Recall(ctx context.Context, conversationID, messageID uuid.UUID) (*Message, error) {
	const q = `
		UPDATE messages
		SET body = '', status = 'recalled', updated_at = $3
		WHERE id = $1 AND conversation_id = $2 AND status IN ('normal', 'edited')
		RETURNING id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
	`
	now := time.Now().UTC()
	row := r.pool.QueryRow(ctx, q, messageID, conversationID, now)
	msg, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return msg, nil
}

type queryRower interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *Repository) getByClientMsgID(ctx context.Context, q queryRower, senderID uuid.UUID, clientMsgID string) (*Message, error) {
	const qry = `
		SELECT id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
		FROM messages
		WHERE sender_id = $1 AND client_msg_id = $2
	`
	row := q.QueryRow(ctx, qry, senderID, clientMsgID)
	return scanMessage(row)
}

type scannable interface {
	Scan(dest ...any) error
}

func scanMessage(row scannable) (*Message, error) {
	var msg Message
	var msgType string
	if err := row.Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.SenderID,
		&msg.ClientMsgID,
		&msg.Seq,
		&msgType,
		&msg.Body,
		&msg.Status,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	); err != nil {
		return nil, err
	}
	msg.Type = Type(msgType)
	return &msg, nil
}
