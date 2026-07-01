package push

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// Provider sends push notifications to devices.
type Provider interface {
	Send(ctx context.Context, platform, token, title, body string) error
}

// MockProvider logs push notifications instead of sending them.
type MockProvider struct {
	logger *slog.Logger
}

// NewMockProvider creates a mock push provider.
func NewMockProvider(logger *slog.Logger) *MockProvider {
	return &MockProvider{logger: logger}
}

// Send logs the push notification.
func (p *MockProvider) Send(ctx context.Context, platform, token, title, body string) error {
	if p.logger != nil {
		p.logger.Info("push mock send",
			"platform", platform,
			"token_prefix", truncate(token, 8),
			"title", title,
			"body", body,
		)
	}
	return nil
}

// Worker dispatches push notifications for offline users.
type Worker struct {
	repo     *Repository
	provider Provider
	logger   *slog.Logger
}

// NewWorker creates a push notification worker.
func NewWorker(repo *Repository, provider Provider, logger *slog.Logger) *Worker {
	return &Worker{repo: repo, provider: provider, logger: logger}
}

// NotifyUser sends a push to all registered tokens for a user.
func (w *Worker) NotifyUser(ctx context.Context, userID uuid.UUID, title, body string) error {
	if w == nil || w.repo == nil || w.provider == nil {
		return nil
	}
	tokens, err := w.repo.List(ctx, userID)
	if err != nil {
		return err
	}
	for _, t := range tokens {
		if err := w.provider.Send(ctx, t.Platform, t.Token, title, body); err != nil && w.logger != nil {
			w.logger.Warn("push send failed", "user_id", userID, "error", err)
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
