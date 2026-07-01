package report

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Report is a message report record.
type Report struct {
	ID             uuid.UUID
	ReporterID     uuid.UUID
	MessageID      uuid.UUID
	ConversationID uuid.UUID
	Reason         string
	CreatedAt      time.Time
}

// Repository manages message reports.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a report repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a message report.
func (r *Repository) Create(ctx context.Context, reporterID, messageID, convID uuid.UUID, reason string) (*Report, error) {
	const q = `
		INSERT INTO message_reports (id, reporter_id, message_id, conversation_id, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, reporter_id, message_id, conversation_id, reason, created_at
	`
	now := time.Now().UTC()
	id := uuid.New()
	row := r.pool.QueryRow(ctx, q, id, reporterID, messageID, convID, reason, now)
	var rpt Report
	if err := row.Scan(&rpt.ID, &rpt.ReporterID, &rpt.MessageID, &rpt.ConversationID, &rpt.Reason, &rpt.CreatedAt); err != nil {
		return nil, fmt.Errorf("create report: %w", err)
	}
	return &rpt, nil
}
