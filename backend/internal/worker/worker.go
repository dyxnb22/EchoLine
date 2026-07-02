package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/echoline/echoline/backend/internal/outbox"
)

const (
	defaultOutboxReclaimInterval   = time.Minute
	defaultOutboxReclaimStaleAfter = 5 * time.Minute
)

// OutboxReclaimWorker resets outbox rows stuck in processing after worker crashes.
type OutboxReclaimWorker struct {
	repo      *outbox.Repository
	staleAfter time.Duration
	logger    *slog.Logger
}

// NewOutboxReclaimWorker creates an outbox reclaim worker.
func NewOutboxReclaimWorker(repo *outbox.Repository, staleAfter time.Duration, logger *slog.Logger) *OutboxReclaimWorker {
	return &OutboxReclaimWorker{repo: repo, staleAfter: staleAfter, logger: logger}
}

// RunOnce reclaims processing rows older than the stale threshold.
func (w *OutboxReclaimWorker) RunOnce(ctx context.Context) (int64, error) {
	if w == nil || w.repo == nil {
		return 0, nil
	}
	staleAfter := w.staleAfter
	if staleAfter <= 0 {
		staleAfter = defaultOutboxReclaimStaleAfter
	}
	return w.repo.ReclaimStaleProcessing(ctx, time.Now().UTC().Add(-staleAfter))
}

// Run periodically reclaims stale processing outbox rows until context cancel.
func (w *OutboxReclaimWorker) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = defaultOutboxReclaimInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := w.RunOnce(ctx)
			if err != nil {
				if w.logger != nil {
					w.logger.Warn("outbox reclaim", "error", err)
				}
				continue
			}
			if n > 0 && w.logger != nil {
				w.logger.Info("outbox reclaimed stale processing", "count", n)
			}
		}
	}
}
