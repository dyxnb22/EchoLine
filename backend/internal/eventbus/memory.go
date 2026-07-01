package eventbus

import "context"

// Event is a domain event published after commit.
type Event struct {
	Type    string
	Payload []byte
}

// Publisher publishes events to async consumers.
type Publisher interface {
	Publish(ctx context.Context, topic string, event Event) error
}

// MemoryPublisher is an in-process publisher for local development.
type MemoryPublisher struct {
	ch chan Event
}

// NewMemoryPublisher creates a buffered in-memory publisher.
func NewMemoryPublisher(buffer int) *MemoryPublisher {
	if buffer <= 0 {
		buffer = 128
	}
	return &MemoryPublisher{ch: make(chan Event, buffer)}
}

// Publish enqueues an event without blocking the write path.
func (p *MemoryPublisher) Publish(ctx context.Context, topic string, event Event) error {
	select {
	case p.ch <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// C returns the consumer channel.
func (p *MemoryPublisher) C() <-chan Event {
	return p.ch
}
