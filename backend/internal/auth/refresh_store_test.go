package auth

import (
	"context"
	"testing"
	"time"
)

func TestMemoryRefreshStoreDetectsReuse(t *testing.T) {
	store := NewMemoryRefreshStore()
	exp := time.Now().UTC().Add(time.Hour)

	reused, err := store.Consume(context.Background(), "jti-1", exp)
	if err != nil || reused {
		t.Fatalf("first consume: reused=%v err=%v", reused, err)
	}

	reused, err = store.Consume(context.Background(), "jti-1", exp)
	if err != nil || !reused {
		t.Fatalf("second consume should detect reuse: reused=%v err=%v", reused, err)
	}
}
