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
	"github.com/echoline/echoline/backend/internal/message"
)

type stubMsgResolver struct {
	convID uuid.UUID
	err    error
}

func (s stubMsgResolver) GetConversationID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	if s.err != nil {
		return uuid.Nil, s.err
	}
	return s.convID, nil
}

type stubMemberChecker struct {
	member bool
	err    error
}

func (s stubMemberChecker) IsMember(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return s.member, s.err
}

func TestReactionServiceRejectsNonMember(t *testing.T) {
	svc := NewReactionService(nil, stubMemberChecker{member: false}, stubMsgResolver{convID: uuid.New()})
	err := svc.Add(context.Background(), uuid.New(), uuid.New(), "👍")
	if err == nil || err.Error() != "not a conversation member" {
		t.Fatalf("Add() error = %v, want not a conversation member", err)
	}
}

func TestReactionServiceRejectsMissingMessage(t *testing.T) {
	svc := NewReactionService(nil, stubMemberChecker{member: true}, stubMsgResolver{err: message.ErrNotFound})
	err := svc.Add(context.Background(), uuid.New(), uuid.New(), "👍")
	if err == nil || err.Error() != "message not found" {
		t.Fatalf("Add() error = %v, want message not found", err)
	}
}

func TestHandleAddReactionNonMemberRejected(t *testing.T) {
	msgID := uuid.New()
	userID := uuid.New()
	svc := NewReactionService(nil, stubMemberChecker{member: false}, stubMsgResolver{convID: uuid.New()})

	h := NewHandler(nil, false)
	h.SetReactionAdder(svc)

	body, _ := json.Marshal(gqlRequest{
		Query:     "mutation addReaction { addReaction }",
		Variables: map[string]any{"messageId": msgID.String(), "emoji": "👍"},
	})
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.Claims{UserID: userID}))
	w := httptest.NewRecorder()
	h.HandleGraphQL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}
