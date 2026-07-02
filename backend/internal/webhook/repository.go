package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Delivery is a persisted webhook delivery attempt.
type Delivery struct {
	ID        uuid.UUID
	EventType string
	Payload   map[string]any
	Status    string
	Attempts  int
	LastError string
	CreatedAt time.Time
}

// Repository persists webhook deliveries for retry.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a webhook delivery repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Enqueue inserts a pending webhook delivery.
func (r *Repository) Enqueue(ctx context.Context, eventType string, payload map[string]any) (uuid.UUID, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal payload: %w", err)
	}
	id := uuid.New()
	const q = `
		INSERT INTO webhook_deliveries (id, event_type, payload, status, attempts, created_at)
		VALUES ($1, $2, $3, 'pending', 0, $4)
	`
	_, err = r.pool.Exec(ctx, q, id, eventType, payloadJSON, time.Now().UTC())
	if err != nil {
		return uuid.Nil, fmt.Errorf("enqueue webhook: %w", err)
	}
	return id, nil
}

// ListPending returns pending deliveries for retry.
func (r *Repository) ListPending(ctx context.Context, limit int) ([]Delivery, error) {
	const q = `
		SELECT id, event_type, payload, status, attempts, COALESCE(last_error, ''), created_at
		FROM webhook_deliveries
		WHERE status = 'pending' AND attempts < 5
		ORDER BY created_at ASC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending webhooks: %w", err)
	}
	defer rows.Close()

	var out []Delivery
	for rows.Next() {
		var d Delivery
		var payloadJSON []byte
		if err := rows.Scan(&d.ID, &d.EventType, &payloadJSON, &d.Status, &d.Attempts, &d.LastError, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan webhook: %w", err)
		}
		_ = json.Unmarshal(payloadJSON, &d.Payload)
		out = append(out, d)
	}
	return out, rows.Err()
}

// MarkDelivered marks a delivery as successful.
func (r *Repository) MarkDelivered(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE webhook_deliveries SET status = 'delivered', delivered_at = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id, time.Now().UTC())
	return err
}

// MarkFailed increments attempts and stores error.
func (r *Repository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	const q = `
		UPDATE webhook_deliveries
		SET attempts = attempts + 1, last_error = $2,
		    status = CASE WHEN attempts + 1 >= 5 THEN 'failed' ELSE 'pending' END
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, q, id, errMsg)
	return err
}

// RetryWorker drains pending webhook deliveries.
type RetryWorker struct {
	repo       *Repository
	dispatcher *Dispatcher
}

// NewRetryWorker creates a webhook retry worker.
func NewRetryWorker(repo *Repository, dispatcher *Dispatcher) *RetryWorker {
	return &RetryWorker{repo: repo, dispatcher: dispatcher}
}

// RunOnce processes one batch of pending deliveries.
func (w *RetryWorker) RunOnce(ctx context.Context) (int, error) {
	if w == nil || w.repo == nil || w.dispatcher == nil || !w.dispatcher.Enabled() {
		return 0, nil
	}
	pending, err := w.repo.ListPending(ctx, 20)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, d := range pending {
		err := w.dispatcher.Dispatch(ctx, d.EventType, d.Payload)
		if err != nil {
			_ = w.repo.MarkFailed(ctx, d.ID, err.Error())
			continue
		}
		_ = w.repo.MarkDelivered(ctx, d.ID)
		processed++
	}
	return processed, nil
}

// PersistingDispatcher enqueues failed webhooks for retry.
type PersistingDispatcher struct {
	inner *Dispatcher
	repo  *Repository
}

// NewPersistingDispatcher wraps a dispatcher with DB persistence on failure.
func NewPersistingDispatcher(inner *Dispatcher, repo *Repository) *PersistingDispatcher {
	return &PersistingDispatcher{inner: inner, repo: repo}
}

// Enabled delegates to inner dispatcher.
func (p *PersistingDispatcher) Enabled() bool {
	return p.inner != nil && p.inner.Enabled()
}

// DispatchMessageCreated fires webhook and enqueues on failure.
func (p *PersistingDispatcher) DispatchMessageCreated(ctx context.Context, messageID, conversationID, senderID string, body string, createdAt time.Time) {
	if p == nil || p.inner == nil || !p.inner.Enabled() {
		return
	}
	payload := map[string]any{
		"message_id":      messageID,
		"conversation_id": conversationID,
		"sender_id":       senderID,
		"body":            body,
		"created_at":      createdAt.UTC().Format(time.RFC3339),
	}
	if err := p.inner.Dispatch(ctx, "message.created", payload); err != nil && p.repo != nil {
		_, _ = p.repo.Enqueue(ctx, "message.created", payload)
	}
}
