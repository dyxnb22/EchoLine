package presence

import (
	"context"
	"time"

	"github.com/echoline/echoline/backend/internal/redisx"
)

const defaultTTL = 90 * time.Second

// Store tracks online presence in Redis with TTL.
type Store struct {
	redis *redisx.Client
	ttl   time.Duration
}

// NewStore creates a Redis-backed presence store.
func NewStore(redis *redisx.Client, ttl time.Duration) *Store {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &Store{redis: redis, ttl: ttl}
}

// Online marks a user/device as online.
func (s *Store) Online(ctx context.Context, userID, deviceID string) error {
	return s.redis.SetPresence(ctx, userID, deviceID, s.ttl)
}

// Offline removes a user/device presence key.
func (s *Store) Offline(ctx context.Context, userID, deviceID string) error {
	return s.redis.DeletePresence(ctx, userID, deviceID)
}

// Refresh extends presence TTL for heartbeat.
func (s *Store) Refresh(ctx context.Context, userID, deviceID string) error {
	return s.Online(ctx, userID, deviceID)
}
