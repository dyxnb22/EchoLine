package push

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestMockProviderSend(t *testing.T) {
	p := NewMockProvider(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err := p.Send(context.Background(), "fcm", "tok123", "title", "body"); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerNotifyUserNilSafe(t *testing.T) {
	w := NewWorker(nil, nil, nil)
	if err := w.NotifyUser(context.Background(), uuid.New(), "t", "b"); err != nil {
		t.Fatal(err)
	}
}
