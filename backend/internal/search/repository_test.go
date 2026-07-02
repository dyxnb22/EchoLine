package search

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestSearchRequiresQuery(t *testing.T) {
	repo := &Repository{}
	_, err := repo.Search(context.Background(), uuid.Nil, "   ", 10)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}
