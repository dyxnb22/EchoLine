package reaction

import (
	"testing"

	"github.com/google/uuid"
)

func TestParseMessageID(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		path    string
		wantErr bool
		wantID  uuid.UUID
	}{
		{"/api/messages/" + id.String() + "/reactions", false, id},
		{"/api/messages/" + id.String() + "/reactions/👍", false, id},
		{"/other/path", true, uuid.Nil},
		{"/api/messages/not-a-uuid/reactions", true, uuid.Nil},
	}

	for _, tc := range tests {
		got, err := parseMessageID(tc.path)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseMessageID(%q): expected error, got nil", tc.path)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseMessageID(%q): unexpected error: %v", tc.path, err)
			continue
		}
		if got != tc.wantID {
			t.Errorf("parseMessageID(%q) = %v; want %v", tc.path, got, tc.wantID)
		}
	}
}

func TestParseMessageIDAndEmoji(t *testing.T) {
	id := uuid.New()
	path := "/api/messages/" + id.String() + "/reactions/👍"
	msgID, emoji, err := parseMessageIDAndEmoji(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgID != id {
		t.Errorf("got message_id %v; want %v", msgID, id)
	}
	if emoji != "👍" {
		t.Errorf("got emoji %q; want %q", emoji, "👍")
	}
}
