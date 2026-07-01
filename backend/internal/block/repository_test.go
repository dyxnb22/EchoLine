package block

import (
	"testing"

	"github.com/google/uuid"
)

func TestBlockSelfError(t *testing.T) {
	r := &Repository{}
	id := uuid.New()
	_, err := r.Block(nil, id, id) //nolint: staticcheck
	if err != ErrSelfBlock {
		t.Fatalf("expected ErrSelfBlock, got %v", err)
	}
}
