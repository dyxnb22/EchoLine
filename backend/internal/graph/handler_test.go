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
)

type stubConvRepo struct{}

func (stubConvRepo) ListForUser(ctx context.Context, userID uuid.UUID, limit int) ([]conversation.Conversation, error) {
	return []conversation.Conversation{{
		ID:    uuid.New(),
		Type:  conversation.TypeDirect,
		Title: "test",
	}}, nil
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
