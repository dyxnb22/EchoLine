package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps a Redis connection used for cache, presence, and rate limiting.
type Client struct {
	rdb *redis.Client
}

// Connect creates a Redis client and verifies connectivity.
func Connect(ctx context.Context, addr string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

// Close closes the Redis client.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Raw returns the underlying go-redis client.
func (c *Client) Raw() *redis.Client {
	return c.rdb
}

// SetPresence stores an online marker with TTL.
func (c *Client) SetPresence(ctx context.Context, userID, deviceID string, ttl time.Duration) error {
	key := fmt.Sprintf("presence:%s:%s", userID, deviceID)
	return c.rdb.Set(ctx, key, "1", ttl).Err()
}

// DeletePresence removes an online marker.
func (c *Client) DeletePresence(ctx context.Context, userID, deviceID string) error {
	key := fmt.Sprintf("presence:%s:%s", userID, deviceID)
	return c.rdb.Del(ctx, key).Err()
}

// HasPresence returns true when at least one presence key exists for the user.
func (c *Client) HasPresence(ctx context.Context, userID string) (bool, error) {
	pattern := fmt.Sprintf("presence:%s:*", userID)
	keys, err := c.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return false, err
	}
	return len(keys) > 0, nil
}

// Allow checks a simple fixed-window rate limit.
func (c *Client) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	count, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		_ = c.rdb.Expire(ctx, key, window).Err()
	}
	return count <= limit, nil
}

// Get returns a string value.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

// Set stores a string value with TTL.
func (c *Client) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Delete removes a key.
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}
