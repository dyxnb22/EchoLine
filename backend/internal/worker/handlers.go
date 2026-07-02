package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/eventbus"
	"github.com/echoline/echoline/backend/internal/metrics"
	"github.com/echoline/echoline/backend/internal/push"
	"github.com/echoline/echoline/backend/internal/search"
)

const maxSeenIDs = 10_000

// MessageCreatedHandler indexes messages and records consumption metrics.
type MessageCreatedHandler struct {
	search *search.Repository
	logger *slog.Logger
	mu     sync.Mutex
	seen   map[string]struct{}
	order  []string
}

// NewMessageCreatedHandler creates a consumer handler.
func NewMessageCreatedHandler(searchRepo *search.Repository, logger *slog.Logger) *MessageCreatedHandler {
	return &MessageCreatedHandler{
		search: searchRepo,
		logger: logger,
		seen:   make(map[string]struct{}),
		order:  make([]string, 0, maxSeenIDs),
	}
}

func (h *MessageCreatedHandler) markSeen(id string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.seen[id]; ok {
		return false
	}
	h.seen[id] = struct{}{}
	h.order = append(h.order, id)
	if len(h.order) > maxSeenIDs {
		oldest := h.order[0]
		h.order = h.order[1:]
		delete(h.seen, oldest)
	}
	return true
}

// Handle processes a message.created payload idempotently.
func (h *MessageCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	evt, err := eventbus.DecodeMessageCreated(payload)
	if err != nil {
		return err
	}
	if !h.markSeen(evt.ID) {
		return nil
	}
	metrics.MQEventsConsumed.WithLabelValues(eventbus.TopicMessageCreated).Inc()

	if h.search == nil {
		return nil
	}

	msgID, err := uuid.Parse(evt.ID)
	if err != nil {
		return err
	}
	convID, err := uuid.Parse(evt.ConversationID)
	if err != nil {
		return err
	}
	senderID, err := uuid.Parse(evt.SenderID)
	if err != nil {
		return err
	}

	createdAt := evt.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return h.search.IndexMessage(ctx, msgID, convID, senderID, evt.Body, evt.Seq, createdAt)
}

// FanoutWorker batches offline notification fanout for large groups.
type FanoutWorker struct {
	pool   *pgxpool.Pool
	push   *push.Worker
	logger *slog.Logger
}

// NewFanoutWorker creates a fanout worker.
func NewFanoutWorker(pool *pgxpool.Pool, pushWorker *push.Worker, logger *slog.Logger) *FanoutWorker {
	return &FanoutWorker{pool: pool, push: pushWorker, logger: logger}
}

const fanoutBatchSize = 256

// Handle notifies offline members in batches (excludes sender).
func (w *FanoutWorker) Handle(ctx context.Context, payload []byte) error {
	evt, err := eventbus.DecodeMessageCreated(payload)
	if err != nil {
		return err
	}
	if w.pool == nil {
		if w.logger != nil {
			w.logger.Info("fanout worker batch", "conversation_id", evt.ConversationID, "seq", evt.Seq)
		}
		return nil
	}

	convID, err := uuid.Parse(evt.ConversationID)
	if err != nil {
		return err
	}
	senderID, err := uuid.Parse(evt.SenderID)
	if err != nil {
		return err
	}

	offset := 0
	notified := 0
	for {
		const q = `
			SELECT user_id FROM conversation_members
			WHERE conversation_id = $1 AND user_id != $2
			ORDER BY user_id
			LIMIT $3 OFFSET $4
		`
		rows, err := w.pool.Query(ctx, q, convID, senderID, fanoutBatchSize, offset)
		if err != nil {
			return err
		}

		batch := 0
		for rows.Next() {
			var uid uuid.UUID
			if err := rows.Scan(&uid); err != nil {
				rows.Close()
				return err
			}
			if w.push != nil {
				_ = w.push.NotifyUser(ctx, uid, "New message", evt.Body)
				notified++
			}
			batch++
		}
		if err := rows.Err(); err != nil {
			return err
		}
		rows.Close()
		if batch < fanoutBatchSize {
			break
		}
		offset += fanoutBatchSize
	}
	if w.logger != nil {
		w.logger.Info("fanout worker batch", "conversation_id", evt.ConversationID, "seq", evt.Seq, "notified", notified)
	}
	return nil
}

// DecodeEvent exposes message.created parsing for worker main.
func DecodeEvent(payload []byte) (eventbus.MessageCreatedEvent, error) {
	return eventbus.DecodeMessageCreated(payload)
}
