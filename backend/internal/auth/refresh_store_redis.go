package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/echoline/echoline/backend/internal/redisx"
)

// RedisRefreshStore tracks consumed refresh JTIs in Redis for multi-instance rotation.
type RedisRefreshStore struct {
	redis *redisx.Client
}

// NewRedisRefreshStore creates a cluster-visible refresh JTI store.
func NewRedisRefreshStore(redis *redisx.Client) *RedisRefreshStore {
	return &RedisRefreshStore{redis: redis}
}

// Consume records a refresh JTI and reports whether it was already used.
func (s *RedisRefreshStore) Consume(ctx context.Context, jti string, expiresAt time.Time) (bool, error) {
	if s == nil || s.redis == nil || jti == "" {
		return false, nil
	}
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return true, nil
	}
	key := fmt.Sprintf("refresh:jti:%s", jti)
	ok, err := s.redis.Raw().SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return false, err
	}
	return !ok, nil
}
