package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Dispatcher fires HTTP webhooks for platform events.
type Dispatcher struct {
	webhookURL string
	httpClient *http.Client
}

// NewDispatcher creates a webhook dispatcher.
// If webhookURL is empty, dispatches are no-ops.
func NewDispatcher(webhookURL string) *Dispatcher {
	return &Dispatcher{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// Enabled returns true when a webhook URL is configured.
func (d *Dispatcher) Enabled() bool {
	return d.webhookURL != ""
}

// Event is a generic webhook payload.
type Event struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
}

// Dispatch sends a webhook event. Failures are logged but not fatal.
func (d *Dispatcher) Dispatch(ctx context.Context, eventType string, payload any) error {
	if !d.Enabled() {
		return nil
	}

	event := Event{
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("webhook marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook dispatch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}
	return nil
}

// DispatchMessageCreated fires a message.created webhook event.
func (d *Dispatcher) DispatchMessageCreated(ctx context.Context, messageID, conversationID, senderID string, body string, createdAt time.Time) {
	if !d.Enabled() {
		return
	}
	_ = d.Dispatch(ctx, "message.created", map[string]any{
		"message_id":      messageID,
		"conversation_id": conversationID,
		"sender_id":       senderID,
		"body":            body,
		"created_at":      createdAt.UTC().Format(time.RFC3339),
	})
}
