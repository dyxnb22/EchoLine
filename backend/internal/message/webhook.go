package message

import (
	"context"
	"time"
)

// WebhookNotifier fires outbound webhooks for message events.
type WebhookNotifier interface {
	DispatchMessageCreated(ctx context.Context, messageID, conversationID, senderID string, body string, createdAt time.Time)
}

// noopWebhook is a no-op notifier.
type noopWebhook struct{}

func (noopWebhook) DispatchMessageCreated(context.Context, string, string, string, string, time.Time) {}

// webhookAdapter adapts the webhook dispatcher interface.
type webhookAdapter struct {
	dispatch func(ctx context.Context, messageID, conversationID, senderID string, body string, createdAt time.Time)
}

func (w webhookAdapter) DispatchMessageCreated(ctx context.Context, messageID, conversationID, senderID string, body string, createdAt time.Time) {
	if w.dispatch != nil {
		w.dispatch(ctx, messageID, conversationID, senderID, body, createdAt)
	}
}

// SetWebhookNotifier configures outbound webhook dispatch on message send.
func (h *Handler) SetWebhookNotifier(n WebhookNotifier) {
	h.webhook = n
}

// notifyWebhook fires message.created webhook asynchronously.
func (h *Handler) notifyWebhook(ctx context.Context, msg *Message) {
	if h.webhook == nil || msg == nil {
		return
	}
	go h.webhook.DispatchMessageCreated(
		context.Background(),
		msg.ID.String(),
		msg.ConversationID.String(),
		msg.SenderID.String(),
		msg.Body,
		msg.CreatedAt,
	)
}

// FuncWebhookNotifier wraps a dispatch function as WebhookNotifier.
func FuncWebhookNotifier(fn func(ctx context.Context, messageID, conversationID, senderID string, body string, createdAt time.Time)) WebhookNotifier {
	return webhookAdapter{dispatch: fn}
}
