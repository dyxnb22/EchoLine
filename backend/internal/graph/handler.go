package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/conversation"
)

// ConversationRepo lists conversations for GraphQL queries.
type ConversationRepo interface {
	ListForUser(ctx context.Context, userID uuid.UUID, limit int) ([]conversation.Conversation, error)
}

// Handler is a minimal GraphQL-style JSON endpoint (prototype).
type Handler struct {
	conversations ConversationRepo
	graphiql      bool
}

// NewHandler creates a GraphQL prototype handler.
func NewHandler(conv ConversationRepo, graphiql bool) *Handler {
	return &Handler{conversations: conv, graphiql: graphiql}
}

type gqlRequest struct {
	Query string `json:"query"`
}

// HandleGraphQL serves POST /graphql with a subset of queries.
func (h *Handler) HandleGraphQL(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && h.graphiql {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(graphiqlHTML))
		return
	}

	if r.Method != http.MethodPost {
		apierror.Write(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "POST required")
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req gqlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	q := strings.ToLower(strings.TrimSpace(req.Query))
	switch {
	case strings.Contains(q, "conversations"):
		h.handleConversations(w, r, claims.UserID)
	default:
		apierror.WriteJSON(w, http.StatusOK, map[string]any{
			"data":   nil,
			"errors": []map[string]string{{"message": "unsupported query; try { conversations { id title } }"}},
		})
	}
}

func (h *Handler) handleConversations(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	if h.conversations == nil {
		apierror.WriteJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"conversations": []any{}}})
		return
	}
	items, err := h.conversations.ListForUser(r.Context(), userID, 50)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list conversations")
		return
	}
	edges := make([]map[string]any, 0, len(items))
	for _, c := range items {
		edges = append(edges, map[string]any{
			"id":    c.ID,
			"title": c.Title,
			"type":  c.Type,
		})
	}
	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{"conversations": edges},
	})
}

const graphiqlHTML = `<!DOCTYPE html><html><head><title>EchoLine GraphiQL</title></head>
<body><h1>EchoLine GraphQL Prototype</h1>
<p>POST JSON: <code>{"query":"{ conversations { id title } }"}</code></p>
</body></html>`
