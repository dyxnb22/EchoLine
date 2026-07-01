package thread

import (
	"testing"

	"github.com/google/uuid"
)

func TestParseThreadPath(t *testing.T) {
	conv := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	msg := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	path := "/api/conversations/" + conv.String() + "/messages/" + msg.String() + "/replies"
	gotConv, gotMsg, err := parseThreadPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if gotConv != conv || gotMsg != msg {
		t.Fatalf("got %v %v", gotConv, gotMsg)
	}
}

func TestParseThreadPathInvalid(t *testing.T) {
	_, _, err := parseThreadPath("/api/messages/x/replies")
	if err == nil {
		t.Fatal("expected error")
	}
}
