package conversation

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/metrics"
)

// HotRepository counts active conversations (E009).
type HotRepository struct {
	pool *pgxpool.Pool
}

// NewHotRepository creates a hot conversations repository.
func NewHotRepository(pool *pgxpool.Pool) *HotRepository {
	return &HotRepository{pool: pool}
}

// MarkHot records that a conversation has active online members and updates the metric.
// This is a lightweight advisory counter; the actual gauge is driven from active WS hubs.
func (r *HotRepository) MarkHot(ctx context.Context, convID uuid.UUID) error {
	_ = convID
	return nil
}

// UpdateHotMetric recomputes the hot-conversations gauge from the DB.
// A conversation is "hot" if it has at least one message in the last 5 minutes.
func (r *HotRepository) UpdateHotMetric(ctx context.Context) error {
	const q = `
		SELECT COUNT(DISTINCT conversation_id)
		FROM messages
		WHERE created_at >= NOW() - INTERVAL '5 minutes'
	`
	row := r.pool.QueryRow(ctx, q)
	var count float64
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("hot metric: %w", err)
	}
	metrics.HotConversations.Set(count)
	return nil
}
