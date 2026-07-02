package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// DeadLetterEvent is a DLQ entry.
type DeadLetterEvent struct {
	ID           uuid.UUID
	SourceTopic  string
	Payload      json.RawMessage
	ErrorMessage string
	Attempts     int
	CreatedAt    time.Time
}

// DLQRepository reads dead letter events.
type DLQRepository struct {
	pool *pgxpool.Pool
}

// NewDLQRepository creates a DLQ reader.
func NewDLQRepository(pool *pgxpool.Pool) *DLQRepository {
	return &DLQRepository{pool: pool}
}

// ListLast50 returns the last 50 dead letter events ordered by created_at DESC.
func (r *DLQRepository) ListLast50(ctx context.Context) ([]DeadLetterEvent, error) {
	const q = `
		SELECT id, source_topic, payload, error_message, attempts, created_at
		FROM dead_letter_events
		ORDER BY created_at DESC
		LIMIT 50
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list dlq: %w", err)
	}
	defer rows.Close()

	var events []DeadLetterEvent
	for rows.Next() {
		var e DeadLetterEvent
		if err := rows.Scan(&e.ID, &e.SourceTopic, &e.Payload, &e.ErrorMessage, &e.Attempts, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan dlq: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// GetByID returns a single dead letter event.
func (r *DLQRepository) GetByID(ctx context.Context, id uuid.UUID) (*DeadLetterEvent, error) {
	const q = `
		SELECT id, source_topic, payload, error_message, attempts, created_at
		FROM dead_letter_events
		WHERE id = $1
	`
	var e DeadLetterEvent
	err := r.pool.QueryRow(ctx, q, id).Scan(&e.ID, &e.SourceTopic, &e.Payload, &e.ErrorMessage, &e.Attempts, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get dlq: %w", err)
	}
	return &e, nil
}

// DLQHandler exposes admin DLQ endpoint.
type DLQHandler struct {
	repo *DLQRepository
}

// NewDLQHandler creates a DLQ handler.
func NewDLQHandler(pool *pgxpool.Pool) *DLQHandler {
	return &DLQHandler{repo: NewDLQRepository(pool)}
}

// HandleList returns the last 50 dead letter events.
// GET /api/admin/dlq
func (h *DLQHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	events, err := h.repo.ListLast50(r.Context())
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list DLQ")
		return
	}

	type item struct {
		ID           uuid.UUID       `json:"id"`
		SourceTopic  string          `json:"source_topic"`
		Payload      json.RawMessage `json:"payload"`
		ErrorMessage string          `json:"error_message"`
		Attempts     int             `json:"attempts"`
		CreatedAt    string          `json:"created_at"`
	}
	out := make([]item, 0, len(events))
	for _, e := range events {
		out = append(out, item{
			ID:           e.ID,
			SourceTopic:  e.SourceTopic,
			Payload:      e.Payload,
			ErrorMessage: e.ErrorMessage,
			Attempts:     e.Attempts,
			CreatedAt:    e.CreatedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"dead_letters": out, "count": len(out)})
}
