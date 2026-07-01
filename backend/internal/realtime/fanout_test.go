package realtime

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/message"
)

func TestBroadcastMessageCreatedExcludesSender(t *testing.T) {
	hub := NewHub()
	sender := uuid.New()
	receiver := uuid.New()

	recvConn := &Connection{UserID: receiver, DeviceID: "d1", send: make(chan []byte, 4)}
	senderConn := &Connection{UserID: sender, DeviceID: "d1", send: make(chan []byte, 4)}
	hub.Register(receiver, "d1", recvConn)
	hub.Register(sender, "d1", senderConn)

	svc := &Server{hub: hub, conversations: &stubConvRepo{members: []uuid.UUID{sender, receiver}}}

	msg := &message.Message{
		ID:             uuid.New(),
		ConversationID: uuid.New(),
		SenderID:       sender,
		Seq:            1,
		Type:           message.TypeText,
		Body:           "hello",
		CreatedAt:      time.Now().UTC(),
	}

	if err := svc.BroadcastMessageCreated(context.Background(), msg.ConversationID, msg, true, sender); err != nil {
		t.Fatalf("broadcast: %v", err)
	}

	select {
	case <-recvConn.send:
	case <-time.After(time.Second):
		t.Fatal("receiver should get push")
	}

	select {
	case <-senderConn.send:
		t.Fatal("sender should be excluded")
	case <-time.After(50 * time.Millisecond):
	}
}

type stubConvRepo struct {
	members []uuid.UUID
}

func (s *stubConvRepo) ListMemberUserIDs(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return s.members, nil
}

func (s *stubConvRepo) IsMember(_ context.Context, _, userID uuid.UUID) (bool, error) {
	for _, id := range s.members {
		if id == userID {
			return true, nil
		}
	}
	return false, nil
}

func (s *stubConvRepo) MarkRead(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int64) error {
	return nil
}
