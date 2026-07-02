package media

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

var (
	ErrAttachmentNotFound = errors.New("attachment not found")
	ErrAttachmentInUse    = errors.New("attachment already linked")
)

// Attachment is stored object metadata.
type Attachment struct {
	ID        uuid.UUID
	MessageID *uuid.UUID
	OwnerID   uuid.UUID
	ObjectKey string
	MimeType  string
	SizeBytes int64
	Checksum  string
	CreatedAt time.Time
}

// Repository persists attachment metadata.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates an attachment repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// RegisterPending records upload metadata before a message references it.
func (r *Repository) RegisterPending(ctx context.Context, ownerID uuid.UUID, objectKey, mimeType string, sizeBytes int64, checksum string) (*Attachment, error) {
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" {
		return nil, fmt.Errorf("object_key is required")
	}
	if !strings.HasPrefix(objectKey, fmt.Sprintf("uploads/%s/", ownerID)) {
		return nil, fmt.Errorf("invalid object_key for owner")
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	id := uuid.New()
	now := time.Now().UTC()
	const q = `
		INSERT INTO attachments (id, message_id, owner_id, object_key, mime_type, size_bytes, checksum, created_at)
		VALUES ($1, NULL, $2, $3, $4, $5, $6, $7)
		RETURNING id, message_id, owner_id, object_key, mime_type, size_bytes, checksum, created_at
	`
	row := r.pool.QueryRow(ctx, q, id, ownerID, objectKey, mimeType, sizeBytes, checksum, now)
	return scanAttachment(row)
}

type execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// GetUnlinkedByObjectKey returns a pending attachment owned by the user.
func (r *Repository) GetUnlinkedByObjectKey(ctx context.Context, ownerID uuid.UUID, objectKey string) (*Attachment, error) {
	const q = `
		SELECT id, message_id, owner_id, object_key, mime_type, size_bytes, checksum, created_at
		FROM attachments
		WHERE owner_id = $1 AND object_key = $2 AND message_id IS NULL
	`
	row := r.pool.QueryRow(ctx, q, ownerID, objectKey)
	att, err := scanAttachment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAttachmentNotFound
		}
		return nil, err
	}
	return att, nil
}

// GetByObjectKey returns an attachment owned by the user (linked or pending).
func (r *Repository) GetByObjectKey(ctx context.Context, ownerID uuid.UUID, objectKey string) (*Attachment, error) {
	const q = `
		SELECT id, message_id, owner_id, object_key, mime_type, size_bytes, checksum, created_at
		FROM attachments
		WHERE owner_id = $1 AND object_key = $2
	`
	row := r.pool.QueryRow(ctx, q, ownerID, objectKey)
	att, err := scanAttachment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAttachmentNotFound
		}
		return nil, err
	}
	return att, nil
}

// GetAccessibleByObjectKey returns an attachment when the user owns it or is a member
// of the conversation that contains the linked message.
func (r *Repository) GetAccessibleByObjectKey(ctx context.Context, userID uuid.UUID, objectKey string) (*Attachment, error) {
	const q = `
		SELECT a.id, a.message_id, a.owner_id, a.object_key, a.mime_type, a.size_bytes, a.checksum, a.created_at
		FROM attachments a
		LEFT JOIN messages m ON m.id = a.message_id
		LEFT JOIN conversation_members cm ON cm.conversation_id = m.conversation_id AND cm.user_id = $1
		WHERE a.object_key = $2
		  AND (a.owner_id = $1 OR cm.user_id IS NOT NULL)
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, q, userID, objectKey)
	att, err := scanAttachment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAttachmentNotFound
		}
		return nil, err
	}
	return att, nil
}

// LinkToMessageInTx associates a pending attachment with a message.
func (r *Repository) LinkToMessageInTx(ctx context.Context, tx execer, attachmentID, messageID uuid.UUID) error {
	const q = `
		UPDATE attachments
		SET message_id = $2
		WHERE id = $1 AND message_id IS NULL
	`
	tag, err := tx.Exec(ctx, q, attachmentID, messageID)
	if err != nil {
		return fmt.Errorf("link attachment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAttachmentInUse
	}
	return nil
}

// ListByMessageIDs returns attachments keyed by message id.
func (r *Repository) ListByMessageIDs(ctx context.Context, messageIDs []uuid.UUID) (map[uuid.UUID]Attachment, error) {
	if len(messageIDs) == 0 {
		return map[uuid.UUID]Attachment{}, nil
	}
	const q = `
		SELECT id, message_id, owner_id, object_key, mime_type, size_bytes, checksum, created_at
		FROM attachments
		WHERE message_id = ANY($1)
	`
	rows, err := r.pool.Query(ctx, q, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("list attachments: %w", err)
	}
	defer rows.Close()

	out := make(map[uuid.UUID]Attachment)
	for rows.Next() {
		att, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		if att.MessageID != nil {
			out[*att.MessageID] = *att
		}
	}
	return out, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAttachment(row scannable) (*Attachment, error) {
	var att Attachment
	if err := row.Scan(
		&att.ID,
		&att.MessageID,
		&att.OwnerID,
		&att.ObjectKey,
		&att.MimeType,
		&att.SizeBytes,
		&att.Checksum,
		&att.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &att, nil
}
