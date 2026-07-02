package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProfileRepository handles profile updates.
type ProfileRepository struct {
	pool *pgxpool.Pool
}

// NewProfileRepository creates a profile repository.
func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

// UpdateDisplayName updates the display_name for a user.
func (r *ProfileRepository) UpdateDisplayName(ctx context.Context, id uuid.UUID, displayName string) error {
	const q = `UPDATE users SET display_name = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, q, strings.TrimSpace(displayName), time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update display_name: %w", err)
	}
	return nil
}
