package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/eventbus"
)

// Publisher drains outbox rows to a bytes publisher.
type Publisher struct {
	outbox   *Repository
	primary  eventbus.BytesPublisher
	fallback eventbus.BytesPublisher
	logger   *slog.Logger
}

// NewPublisher creates an outbox drainer.
func NewPublisher(outbox *Repository, primary, fallback eventbus.BytesPublisher, logger *slog.Logger) *Publisher {
	return &Publisher{
		outbox:   outbox,
		primary:  primary,
		fallback: fallback,
		logger:   logger,
	}
}

// Run polls and publishes pending events until context cancel.
func (p *Publisher) Run(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.flush(ctx)
		}
	}
}

func (p *Publisher) flush(ctx context.Context) {
	events, err := p.outbox.FetchPending(ctx, 50)
	if err != nil {
		p.logger.Error("fetch outbox", "error", err)
		return
	}
	for _, evt := range events {
		pub := p.primary
		if pub == nil {
			pub = p.fallback
		}
		if pub == nil {
			return
		}
		if err := pub.Publish(ctx, evt.Topic, evt.Payload); err != nil {
			p.logger.Error("publish outbox event", "id", evt.ID, "error", err)
			if p.fallback != nil && pub != p.fallback {
				if err2 := p.fallback.Publish(ctx, evt.Topic, evt.Payload); err2 == nil {
					p.markPublishedWithRetry(ctx, evt.ID)
					continue
				}
			}
			_ = p.outbox.MarkFailed(ctx, evt.ID)
			continue
		}
		p.markPublishedWithRetry(ctx, evt.ID)
	}
}

func (p *Publisher) markPublishedWithRetry(ctx context.Context, id uuid.UUID) {
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if err := p.outbox.MarkPublished(ctx, id); err == nil {
			return
		} else {
			lastErr = err
		}
		time.Sleep(time.Duration(attempt+1) * 50 * time.Millisecond)
	}
	if lastErr != nil {
		p.logger.Error("mark outbox published", "id", id, "error", lastErr)
	}
}
