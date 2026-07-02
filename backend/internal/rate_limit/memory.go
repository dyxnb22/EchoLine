package rate_limit

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/echoline/echoline/backend/internal/auth"
)

type memoryEntry struct {
	count int64
	reset time.Time
}

// MemoryLimiter provides in-process fixed-window rate limiting.
type MemoryLimiter struct {
	mu    sync.Mutex
	slots map[string]memoryEntry
}

// NewMemoryLimiter creates an in-memory limiter used when Redis is unavailable.
func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{slots: make(map[string]memoryEntry)}
}

// Allow implements Limiter.
func (m *MemoryLimiter) Allow(_ context.Context, key string, limit int64, window time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	entry, ok := m.slots[key]
	if !ok || now.After(entry.reset) {
		m.slots[key] = memoryEntry{count: 1, reset: now.Add(window)}
		return true, nil
	}
	if entry.count >= limit {
		return false, nil
	}
	entry.count++
	m.slots[key] = entry
	return true, nil
}

// AuthUserKey returns the authenticated user ID for per-user limits.
func AuthUserKey(r *http.Request) string {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		return IPKey(r)
	}
	return claims.UserID.String()
}
