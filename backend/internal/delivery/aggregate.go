package delivery

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AggregateStatus is the computed delivery state across all devices.
type AggregateStatus struct {
	MessageID uuid.UUID
	Status    Status
}

// Aggregator computes the aggregate delivery/read status across devices.
type Aggregator struct {
	pool *pgxpool.Pool
}

// NewAggregator creates a delivery aggregator.
func NewAggregator(pool *pgxpool.Pool) *Aggregator {
	return &Aggregator{pool: pool}
}

// AggregateReadStatus returns the minimum (worst) delivery status across all devices for a user+message.
// A message is "read" only if at least one device has read it.
// A message is "delivered" if at least one device has received it and none have read it.
// Otherwise "sent".
func (a *Aggregator) AggregateReadStatus(ctx context.Context, messageID, userID uuid.UUID) (*AggregateStatus, error) {
	const q = `
		SELECT status
		FROM message_deliveries
		WHERE message_id = $1 AND user_id = $2
	`
	rows, err := a.pool.Query(ctx, q, messageID, userID)
	if err != nil {
		return nil, fmt.Errorf("aggregate read status: %w", err)
	}
	defer rows.Close()

	best := StatusSent
	for rows.Next() {
		var s Status
		if err := rows.Scan(&s); err != nil {
			return nil, fmt.Errorf("scan status: %w", err)
		}
		if statusRank[s] > statusRank[best] {
			best = s
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &AggregateStatus{MessageID: messageID, Status: best}, nil
}
