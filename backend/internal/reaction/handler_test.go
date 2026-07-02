package reaction

import (
	"testing"

	"github.com/google/uuid"
)

func TestParseMessageID(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	got, err := parseMessageID("/api/messages/" + id.String() + "/reactions")
	if err != nil {
		t.Fatal(err)
	}
	if got != id {
		t.Fatalf("got %v want %v", got, id)
	}
}

func TestParseMessageIDAndEmoji(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	gotID, emoji, err := parseMessageIDAndEmoji("/api/messages/" + id.String() + "/reactions/👍")
	if err != nil {
		t.Fatal(err)
	}
	if gotID != id || emoji != "👍" {
		t.Fatalf("got %v %q", gotID, emoji)
	}
}

func TestParseMessageIDInvalid(t *testing.T) {
	_, err := parseMessageID("/api/conversations/x")
	if err == nil {
		t.Fatal("expected error")
	}
}
