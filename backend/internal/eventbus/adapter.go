package eventbus

import "context"

// BytesPublisher adapts publishers that accept raw payloads.
type BytesPublisher interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

// MemoryBytesPublisher adapts MemoryPublisher to BytesPublisher.
type MemoryBytesPublisher struct {
	inner *MemoryPublisher
}

// NewMemoryBytesPublisher wraps a memory publisher.
func NewMemoryBytesPublisher(inner *MemoryPublisher) *MemoryBytesPublisher {
	return &MemoryBytesPublisher{inner: inner}
}

// Publish implements BytesPublisher.
func (p *MemoryBytesPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
	return p.inner.Publish(ctx, topic, Event{Type: topic, Payload: payload})
}

// FanoutPublisher tries primary then falls back on error.
type FanoutPublisher struct {
	primary  BytesPublisher
	fallback BytesPublisher
}

// NewFanoutPublisher creates a chained publisher.
func NewFanoutPublisher(primary, fallback BytesPublisher) *FanoutPublisher {
	return &FanoutPublisher{primary: primary, fallback: fallback}
}

// Publish tries primary publisher, then fallback.
func (p *FanoutPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
	if p.primary != nil {
		if err := p.primary.Publish(ctx, topic, payload); err == nil {
			return nil
		}
	}
	if p.fallback != nil {
		return p.fallback.Publish(ctx, topic, payload)
	}
	return nil
}
