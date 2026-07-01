package entitlement

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository manages channel access entitlements.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates an entitlement repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// RequiresEntitlement returns true when the channel requires a paid entitlement.
func (r *Repository) RequiresEntitlement(ctx context.Context, channelID uuid.UUID) (bool, error) {
	const q = `SELECT requires_entitlement FROM conversations WHERE id = $1 AND type = 'channel'`
	var required bool
	err := r.pool.QueryRow(ctx, q, channelID).Scan(&required)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("requires entitlement: %w", err)
	}
	return required, nil
}

// HasActiveEntitlement checks whether the user may access a paid channel.
func (r *Repository) HasActiveEntitlement(ctx context.Context, userID, channelID uuid.UUID) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1 FROM channel_entitlements
			WHERE user_id = $1 AND channel_id = $2 AND status = 'active'
			  AND (expires_at IS NULL OR expires_at > NOW())
		)
	`
	var ok bool
	if err := r.pool.QueryRow(ctx, q, userID, channelID).Scan(&ok); err != nil {
		return false, fmt.Errorf("has entitlement: %w", err)
	}
	return ok, nil
}

// Grant creates or renews an active entitlement (idempotent by reference).
func (r *Repository) Grant(ctx context.Context, userID, channelID uuid.UUID, reference string) error {
	const q = `
		INSERT INTO channel_entitlements (id, user_id, channel_id, status, reference, created_at)
		VALUES (gen_random_uuid(), $1, $2, 'active', $3, $4)
		ON CONFLICT (user_id, channel_id) DO UPDATE
			SET status = 'active', reference = EXCLUDED.reference, created_at = EXCLUDED.created_at
	`
	_, err := r.pool.Exec(ctx, q, userID, channelID, reference, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("grant entitlement: %w", err)
	}
	return nil
}

// SetChannelRequiresEntitlement marks a channel as paid.
func (r *Repository) SetChannelRequiresEntitlement(ctx context.Context, channelID uuid.UUID, required bool) error {
	const q = `UPDATE conversations SET requires_entitlement = $2 WHERE id = $1 AND type = 'channel'`
	tag, err := r.pool.Exec(ctx, q, channelID, required)
	if err != nil {
		return fmt.Errorf("set requires entitlement: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotPaidChannel
	}
	return nil
}

// CanSubscribe returns nil when the user may subscribe to the channel.
func (r *Repository) CanSubscribe(ctx context.Context, userID, channelID uuid.UUID) error {
	required, err := r.RequiresEntitlement(ctx, channelID)
	if err != nil {
		return err
	}
	if !required {
		return nil
	}
	ok, err := r.HasActiveEntitlement(ctx, userID, channelID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrEntitlementRequired
	}
	return nil
}
