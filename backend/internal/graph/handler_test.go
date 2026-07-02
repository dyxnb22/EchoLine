package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/message"
)

type stubConvRepo struct{}

func (stubConvRepo) ListForUser(ctx context.Context, userID uuid.UUID, limit int) ([]conversation.Conversation, error) {
	return []conversation.Conversation{{
		ID:    uuid.New(),
		Type:  conversation.TypeDirect,
		Title: "test",
	}}, nil
}

type stubMessageSender struct {
	gotClientMsgID string
}

func (s *stubMessageSender) Send(ctx context.Context, convID, userID uuid.UUID, input message.SendInput) (*message.Message, error) {
	s.gotClientMsgID = input.ClientMsgID
	return &message.Message{ID: uuid.New(), Seq: 1, Body: input.Body}, nil
}

func TestHandleGraphQLConversations(t *testing.T) {
	h := NewHandler(stubConvRepo{}, false)
	body, _ := json.Marshal(gqlRequest{Query: "{ conversations { id title } }"})
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.Claims{UserID: uuid.New()}))
	w := httptest.NewRecorder()
	h.HandleGraphQL(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestHandleGraphQLSendMessageSetsClientMsgID(t *testing.T) {
	sender := &stubMessageSender{}
	h := NewHandler(stubConvRepo{}, false)
	h.SetMessageSender(sender)
	convID := uuid.New()
	body, _ := json.Marshal(gqlRequest{
		Query:     "mutation { sendMessage }",
		Variables: map[string]any{"conversationId": convID.String(), "body": "hi"},
	})
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.Claims{UserID: uuid.New()}))
	w := httptest.NewRecorder()
	h.HandleGraphQL(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if sender.gotClientMsgID == "" {
		t.Fatal("expected non-empty client_msg_id for GraphQL sendMessage")
	}
}
