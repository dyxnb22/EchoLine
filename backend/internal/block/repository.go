package block

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAlreadyBlocked = errors.New("user already blocked")
var ErrNotBlocked = errors.New("user is not blocked")
var ErrSelfBlock = errors.New("cannot block yourself")

// Block represents a user block relationship.
type Block struct {
	BlockerID uuid.UUID
	BlockedID uuid.UUID
	CreatedAt time.Time
}

// Repository manages user blocks.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a block repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Block creates a block from blockerID -> blockedID.
func (r *Repository) Block(ctx context.Context, blockerID, blockedID uuid.UUID) (*Block, error) {
	if blockerID == blockedID {
		return nil, ErrSelfBlock
	}
	const q = `
		INSERT INTO user_blocks (blocker_id, blocked_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (blocker_id, blocked_id) DO NOTHING
		RETURNING blocker_id, blocked_id, created_at
	`
	now := time.Now().UTC()
	row := r.pool.QueryRow(ctx, q, blockerID, blockedID, now)
	var b Block
	if err := row.Scan(&b.BlockerID, &b.BlockedID, &b.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAlreadyBlocked
		}
		return nil, fmt.Errorf("block user: %w", err)
	}
	return &b, nil
}

// Unblock removes a block.
func (r *Repository) Unblock(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	const q = `DELETE FROM user_blocks WHERE blocker_id = $1 AND blocked_id = $2`
	tag, err := r.pool.Exec(ctx, q, blockerID, blockedID)
	if err != nil {
		return fmt.Errorf("unblock user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotBlocked
	}
	return nil
}

// List returns all users blocked by blockerID.
func (r *Repository) List(ctx context.Context, blockerID uuid.UUID) ([]Block, error) {
	const q = `
		SELECT blocker_id, blocked_id, created_at
		FROM user_blocks
		WHERE blocker_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, blockerID)
	if err != nil {
		return nil, fmt.Errorf("list blocks: %w", err)
	}
	defer rows.Close()

	var blocks []Block
	for rows.Next() {
		var b Block
		if err := rows.Scan(&b.BlockerID, &b.BlockedID, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan block: %w", err)
		}
		blocks = append(blocks, b)
	}
	return blocks, rows.Err()
}

// IsBlocked returns true if blockerID has blocked blockedID.
func (r *Repository) IsBlocked(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error) {
	const q = `SELECT 1 FROM user_blocks WHERE blocker_id = $1 AND blocked_id = $2`
	row := r.pool.QueryRow(ctx, q, blockerID, blockedID)
	var dummy int
	err := row.Scan(&dummy)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("is blocked: %w", err)
	}
	return true, nil
}
