package conversation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetMember returns membership details for a user.
func (r *Repository) GetMember(ctx context.Context, conversationID, userID uuid.UUID) (*Member, error) {
	const q = `
		SELECT conversation_id, user_id, role, last_read_seq, joined_at
		FROM conversation_members
		WHERE conversation_id = $1 AND user_id = $2
	`
	var m Member
	var role string
	err := r.pool.QueryRow(ctx, q, conversationID, userID).Scan(
		&m.ConversationID, &m.UserID, &role, &m.LastReadSeq, &m.JoinedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotMember
	}
	if err != nil {
		return nil, fmt.Errorf("get member: %w", err)
	}
	m.Role = Role(role)
	return &m, nil
}

// CanPublish checks whether a user may send messages in a conversation.
func (r *Repository) CanPublish(ctx context.Context, conversationID, userID uuid.UUID) error {
	conv, err := r.GetByID(ctx, conversationID)
	if err != nil {
		return err
	}
	member, err := r.GetMember(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if !CanPublish(conv.Type, member.Role) {
		return ErrCannotPublish
	}
	return nil
}

// CreateChannel creates a broadcast channel owned by the creator.
func (r *Repository) CreateChannel(ctx context.Context, ownerID uuid.UUID, title string) (*Conversation, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("channel title is required")
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
		VALUES ($1, 'channel', $2, 0, $3, $4, $5)
		RETURNING id, type, title, latest_seq, last_message_id, created_by, created_at, updated_at
	`
	row := tx.QueryRow(ctx, insertConv, convID, title, ownerID, now, now)
	conv, err := scanConversation(row)
	if err != nil {
		return nil, fmt.Errorf("insert channel: %w", err)
	}

	if err := insertMember(ctx, tx, conv.ID, ownerID, RoleOwner, now); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit channel: %w", err)
	}
	return conv, nil
}

// Subscribe adds a subscriber to a channel.
func (r *Repository) Subscribe(ctx context.Context, channelID, userID uuid.UUID) error {
	conv, err := r.GetByID(ctx, channelID)
	if err != nil {
		return err
	}
	if conv.Type != TypeChannel {
		return ErrInvalidType
	}

	now := time.Now().UTC()
	const q = `
		INSERT INTO conversation_members (conversation_id, user_id, role, joined_at)
		VALUES ($1, $2, 'subscriber', $3)
		ON CONFLICT (conversation_id, user_id) DO NOTHING
	`
	if _, err := r.pool.Exec(ctx, q, channelID, userID, now); err != nil {
		return fmt.Errorf("subscribe channel: %w", err)
	}
	return nil
}

// Unsubscribe removes a user from a channel.
func (r *Repository) Unsubscribe(ctx context.Context, channelID, userID uuid.UUID) error {
	conv, err := r.GetByID(ctx, channelID)
	if err != nil {
		return err
	}
	if conv.Type != TypeChannel {
		return ErrInvalidType
	}

	const q = `DELETE FROM conversation_members WHERE conversation_id = $1 AND user_id = $2 AND role = 'subscriber'`
	tag, err := r.pool.Exec(ctx, q, channelID, userID)
	if err != nil {
		return fmt.Errorf("unsubscribe channel: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotMember
	}
	return nil
}

// AddMember invites a user to a group with the given role.
func (r *Repository) AddMember(ctx context.Context, conversationID, userID uuid.UUID, role Role) error {
	conv, err := r.GetByID(ctx, conversationID)
	if err != nil {
		return err
	}
	if conv.Type != TypeGroup {
		return ErrInvalidType
	}
	if role != RoleMember && role != RoleAdmin {
		return fmt.Errorf("invalid role for invite")
	}

	now := time.Now().UTC()
	const q = `
		INSERT INTO conversation_members (conversation_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (conversation_id, user_id) DO NOTHING
	`
	if _, err := r.pool.Exec(ctx, q, conversationID, userID, string(role), now); err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

// RemoveMember removes a user from a group conversation.
func (r *Repository) RemoveMember(ctx context.Context, conversationID, userID uuid.UUID) error {
	conv, err := r.GetByID(ctx, conversationID)
	if err != nil {
		return err
	}
	if conv.Type != TypeGroup {
		return ErrInvalidType
	}

	member, err := r.GetMember(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if member.Role == RoleOwner {
		return ErrForbidden
	}

	const q = `DELETE FROM conversation_members WHERE conversation_id = $1 AND user_id = $2`
	tag, err := r.pool.Exec(ctx, q, conversationID, userID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotMember
	}
	return nil
}
