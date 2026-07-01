package validate

import (
	"strings"
	"testing"
)

func TestUsername(t *testing.T) {
	if _, err := Username(""); err != ErrUsernameEmpty {
		t.Fatalf("empty username: %v", err)
	}
	long := strings.Repeat("a", MaxUsernameLen+1)
	if _, err := Username(long); err != ErrUsernameTooLong {
		t.Fatalf("long username: %v", err)
	}
	got, err := Username("  alice  ")
	if err != nil || got != "alice" {
		t.Fatalf("trim: got %q err %v", got, err)
	}
}

func TestMessageBody(t *testing.T) {
	if err := MessageBody("", false); err != ErrMessageBodyEmpty {
		t.Fatalf("empty body: %v", err)
	}
	if err := MessageBody("", true); err != nil {
		t.Fatalf("attachment only: %v", err)
	}
	long := strings.Repeat("x", MaxMessageBodyLen+1)
	if err := MessageBody(long, false); err != ErrMessageBodyLong {
		t.Fatalf("long body: %v", err)
	}
}
