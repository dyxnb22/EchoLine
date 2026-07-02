package auth

import (
	"context"
	"sync"
	"time"
)

// RefreshStore tracks consumed refresh token JTIs to enable rotation.
type RefreshStore interface {
	Consume(ctx context.Context, jti string, expiresAt time.Time) (reused bool, err error)
}

type memoryRefreshStore struct {
	mu    sync.Mutex
	used  map[string]time.Time
}

// NewMemoryRefreshStore creates an in-process refresh JTI store.
func NewMemoryRefreshStore() RefreshStore {
	return &memoryRefreshStore{used: make(map[string]time.Time)}
}

func (s *memoryRefreshStore) Consume(_ context.Context, jti string, expiresAt time.Time) (bool, error) {
	if jti == "" {
		return false, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	for id, exp := range s.used {
		if exp.Before(now) {
			delete(s.used, id)
		}
	}
	if exp, ok := s.used[jti]; ok && exp.After(now) {
		return true, nil
	}
	s.used[jti] = expiresAt
	return false, nil
}
