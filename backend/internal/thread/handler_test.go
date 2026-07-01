package thread

import (
	"testing"

	"github.com/google/uuid"
)

func TestParseThreadPath(t *testing.T) {
	convID := uuid.New()
	msgID := uuid.New()
	path := "/api/conversations/" + convID.String() + "/messages/" + msgID.String() + "/replies"

	gotConvID, gotMsgID, err := parseThreadPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotConvID != convID {
		t.Errorf("got conv_id %v; want %v", gotConvID, convID)
	}
	if gotMsgID != msgID {
		t.Errorf("got msg_id %v; want %v", gotMsgID, msgID)
	}

	// invalid paths
	badPaths := []string{
		"/other/path",
		"/api/conversations/" + convID.String() + "/not-messages/replies",
		"/api/conversations/not-uuid/messages/" + msgID.String() + "/replies",
	}
	for _, p := range badPaths {
		_, _, err := parseThreadPath(p)
		if err == nil {
			t.Errorf("parseThreadPath(%q): expected error, got nil", p)
		}
	}
}
