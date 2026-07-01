package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/eventbus"
	"github.com/echoline/echoline/backend/internal/metrics"
	"github.com/echoline/echoline/backend/internal/search"
)

// MessageCreatedHandler indexes messages and records consumption metrics.
type MessageCreatedHandler struct {
	search *search.Repository
	logger *slog.Logger
	seen   map[string]struct{}
}

// NewMessageCreatedHandler creates a consumer handler.
func NewMessageCreatedHandler(searchRepo *search.Repository, logger *slog.Logger) *MessageCreatedHandler {
	return &MessageCreatedHandler{
		search: searchRepo,
		logger: logger,
		seen:   make(map[string]struct{}),
	}
}

// Handle processes a message.created payload idempotently.
func (h *MessageCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	evt, err := eventbus.DecodeMessageCreated(payload)
	if err != nil {
		return err
	}
	if _, ok := h.seen[evt.ID]; ok {
		return nil
	}
	h.seen[evt.ID] = struct{}{}
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

// FanoutWorker batches online fanout work from consumed events (skeleton).
type FanoutWorker struct {
	logger *slog.Logger
}

// NewFanoutWorker creates a fanout worker skeleton.
func NewFanoutWorker(logger *slog.Logger) *FanoutWorker {
	return &FanoutWorker{logger: logger}
}

// Handle logs fanout intent for large groups; hot path still uses WS hub.
func (w *FanoutWorker) Handle(ctx context.Context, payload []byte) error {
	evt, err := eventbus.DecodeMessageCreated(payload)
	if err != nil {
		return err
	}
	w.logger.Info("fanout worker batch", "conversation_id", evt.ConversationID, "seq", evt.Seq)
	return nil
}
