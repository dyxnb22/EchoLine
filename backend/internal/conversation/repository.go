package conversation

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

// Repository persists conversations and members.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a conversation repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateDirect finds or creates a direct conversation between two users.
func (r *Repository) CreateDirect(ctx context.Context, userA, userB uuid.UUID) (*Conversation, bool, error) {
	if userA == userB {
		return nil, false, fmt.Errorf("cannot create direct conversation with self")
	}

	low, high := userA, userB
	if low.String() > high.String() {
		low, high = high, low
	}

	const findQ = `
		SELECT c.id, c.type, c.title, c.latest_seq, c.last_message_id, c.created_by, c.created_at, c.updated_at
		FROM direct_conversation_pairs p
		JOIN conversations c ON c.id = p.conversation_id
		WHERE p.user_low = $1 AND p.user_high = $2
	`
	row := r.pool.QueryRow(ctx, findQ, low, high)
	existing, err := scanConversation(row)
	if err == nil {
		return existing, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, fmt.Errorf("find direct conversation: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	convID := uuid.New()

	const insertConv = `
		INSERT INTO conversations (id, type, title, latest_seq, created_by, created_at, updated_at)
		VALUES ($1, 'direct', '', 0, $2, $3, $4)
		RETURNING id, type, title, latest_seq, last_message_id, created_by, created_at, updated_at
	`
	row = tx.QueryRow(ctx, insertConv, convID, userA, now, now)
	conv, err := scanConversation(row)
	if err != nil {
		return nil, false, fmt.Errorf("insert direct conversation: %w", err)
	}

	for _, member := range []struct {
		userID uuid.UUID
		role   Role
	}{
		{userA, RoleMember},
		{userB, RoleMember},
	} {
		if err := insertMember(ctx, tx, conv.ID, member.userID, member.role, now); err != nil {
			return nil, false, err
		}
	}

	const insertPair = `
		INSERT INTO direct_conversation_pairs (user_low, user_high, conversation_id)
		VALUES ($1, $2, $3)
	`
	if _, err := tx.Exec(ctx, insertPair, low, high, conv.ID); err != nil {
		return nil, false, fmt.Errorf("insert direct pair: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, fmt.Errorf("commit direct conversation: %w", err)
	}

	return conv, true, nil
}

// CreateGroup creates a new group conversation.
func (r *Repository) CreateGroup(ctx context.Context, ownerID uuid.UUID, title string, memberIDs []uuid.UUID) (*Conversation, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("group title is required")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	convID := uuid.New()

	const insertConv = `
		INSERT INTO conversations (id, type, title, latest_seq, created_by, created_at, updated_at)
		VALUES ($1, 'group', $2, 0, $3, $4, $5)
		RETURNING id, type, title, latest_seq, last_message_id, created_by, created_at, updated_at
	`
	row := tx.QueryRow(ctx, insertConv, convID, title, ownerID, now, now)
	conv, err := scanConversation(row)
	if err != nil {
		return nil, fmt.Errorf("insert group conversation: %w", err)
	}

	if err := insertMember(ctx, tx, conv.ID, ownerID, RoleOwner, now); err != nil {
		return nil, err
	}

	seen := map[uuid.UUID]struct{}{ownerID: {}}
	for _, memberID := range memberIDs {
		if memberID == ownerID {
			continue
		}
		if _, ok := seen[memberID]; ok {
			continue
		}
		seen[memberID] = struct{}{}
		if err := insertMember(ctx, tx, conv.ID, memberID, RoleMember, now); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit group conversation: %w", err)
	}

	return conv, nil
}

// ListForUser returns conversations the user belongs to, ordered by recent activity.
func (r *Repository) ListForUser(ctx context.Context, userID uuid.UUID, limit int) ([]Conversation, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	const q = `
		SELECT c.id, c.type, c.title, c.latest_seq, c.last_message_id, c.created_by, c.created_at, c.updated_at
		FROM conversations c
		INNER JOIN conversation_members m ON m.conversation_id = c.id
		WHERE m.user_id = $1
		ORDER BY c.updated_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		conv, err := scanConversation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}
		conversations = append(conversations, *conv)
	}
	return conversations, rows.Err()
}

// GetByID loads a conversation by ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Conversation, error) {
	const q = `
		SELECT id, type, title, latest_seq, last_message_id, created_by, created_at, updated_at
		FROM conversations
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, q, id)
	conv, err := scanConversation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get conversation: %w", err)
	}
	return conv, nil
}

// IsMember checks whether a user belongs to a conversation.
func (r *Repository) IsMember(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	const q = `
		SELECT 1
		FROM conversation_members
		WHERE conversation_id = $1 AND user_id = $2
	`
	var one int
	err := r.pool.QueryRow(ctx, q, conversationID, userID).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check membership: %w", err)
	}
	return true, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanConversation(row scannable) (*Conversation, error) {
	var conv Conversation
	var convType string
	if err := row.Scan(
		&conv.ID,
		&convType,
		&conv.Title,
		&conv.LatestSeq,
		&conv.LastMessageID,
		&conv.CreatedBy,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	); err != nil {
		return nil, err
	}
	conv.Type = Type(convType)
	return &conv, nil
}

type execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func insertMember(ctx context.Context, tx execer, conversationID, userID uuid.UUID, role Role, joinedAt time.Time) error {
	const q = `
		INSERT INTO conversation_members (conversation_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := tx.Exec(ctx, q, conversationID, userID, string(role), joinedAt); err != nil {
		return fmt.Errorf("insert member: %w", err)
	}
	return nil
}
