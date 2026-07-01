package risk

import (
	"testing"

	"github.com/google/uuid"
)

func TestCheckDuplicateBody_NoDuplicate(t *testing.T) {
	s := NewSpamChecker()
	id := uuid.New()
	if err := s.CheckDuplicateBody(id, "hello world"); err != nil {
		t.Fatalf("unexpected error for first send: %v", err)
	}
	if err := s.CheckDuplicateBody(id, "different message"); err != nil {
		t.Fatalf("unexpected error for different body: %v", err)
	}
}

func TestCheckDuplicateBody_SpamDetected(t *testing.T) {
	s := NewSpamChecker()
	s.maxRep = 2
	id := uuid.New()

	for i := 0; i < 2; i++ {
		if err := s.CheckDuplicateBody(id, "spam"); err != nil {
			t.Fatalf("unexpected error on send %d: %v", i+1, err)
		}
	}

	if err := s.CheckDuplicateBody(id, "spam"); err == nil {
		t.Fatal("expected ErrSpamDetected, got nil")
	}
}

func TestCheckDuplicateBody_DifferentUsers(t *testing.T) {
	s := NewSpamChecker()
	s.maxRep = 1
	userA := uuid.New()
	userB := uuid.New()

	_ = s.CheckDuplicateBody(userA, "hello")

	// Same body but different user should not trigger spam for userB
	if err := s.CheckDuplicateBody(userB, "hello"); err != nil {
		t.Fatalf("different user should not trigger spam: %v", err)
	}
}

func TestCheckDuplicateBody_KeyIsolation(t *testing.T) {
	s := NewSpamChecker()
	s.maxRep = 2
	id := uuid.New()

	for i := 0; i < 2; i++ {
		_ = s.CheckDuplicateBody(id, "spam")
	}

	// Different body should NOT be flagged
	if err := s.CheckDuplicateBody(id, "different"); err != nil {
		t.Fatalf("different body should not be spam: %v", err)
	}
}
