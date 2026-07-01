package message

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists messages.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a message repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a message and allocates the next conversation seq in one transaction.
func (r *Repository) Create(ctx context.Context, conversationID, senderID uuid.UUID, clientMsgID string, msgType Type, body string) (*Message, error) {
	clientMsgID = strings.TrimSpace(clientMsgID)
	body = strings.TrimSpace(body)
	if body == "" {
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
