package admin

import (
	"testing"

	"github.com/google/uuid"
)

func TestStaticAdminChecker(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	c := NewStaticAdminChecker(id.String())
	if !c.IsAdmin(id) {
		t.Fatal("expected admin")
	}
	if c.IsAdmin(uuid.New()) {
		t.Fatal("expected non-admin")
	}
}

func TestStaticAdminCheckerEmpty(t *testing.T) {
	c := NewStaticAdminChecker("")
	if c.IsAdmin(uuid.New()) {
		t.Fatal("empty checker should deny all")
	}
}
