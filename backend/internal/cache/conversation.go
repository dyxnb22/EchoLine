package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/echoline/echoline/backend/internal/redisx"
)

const convListTTL = 30 * time.Second

// ConversationSummary is a cached conversation list item.
type ConversationSummary struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Unread    int64  `json:"unread"`
	LatestSeq int64  `json:"latest_seq"`
}

// ConversationListCache caches per-user conversation summaries in Redis.
type ConversationListCache struct {
	redis *redisx.Client
}

// NewConversationListCache creates a cache helper.
func NewConversationListCache(redis *redisx.Client) *ConversationListCache {
	if redis == nil {
		return nil
	}
	return &ConversationListCache{redis: redis}
}

func convListKey(userID string) string {
	return fmt.Sprintf("conv:list:%s", userID)
}

// Get returns cached summaries if present.
func (c *ConversationListCache) Get(ctx context.Context, userID string) ([]ConversationSummary, bool, error) {
	if c == nil || c.redis == nil {
		return nil, false, nil
	}
	raw, err := c.redis.Get(ctx, convListKey(userID))
	if err != nil || raw == "" {
		return nil, false, err
	}
	var items []ConversationSummary
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, false, err
	}
	return items, true, nil
}

// Set stores conversation summaries with TTL.
func (c *ConversationListCache) Set(ctx context.Context, userID string, items []ConversationSummary) error {
	if c == nil || c.redis == nil {
		return nil
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, convListKey(userID), string(raw), convListTTL)
}

// Invalidate removes cached list for a user.
func (c *ConversationListCache) Invalidate(ctx context.Context, userID string) error {
	if c == nil || c.redis == nil {
		return nil
	}
	return c.redis.Delete(ctx, convListKey(userID))
}
