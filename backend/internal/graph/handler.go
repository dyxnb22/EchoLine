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
	"github.com/echoline/echoline/backend/internal/message"
)

// ConversationRepo lists conversations for GraphQL queries.
type ConversationRepo interface {
	ListForUser(ctx context.Context, userID uuid.UUID, limit int) ([]conversation.Conversation, error)
}

// MessageSender sends messages for GraphQL mutations.
type MessageSender interface {
	Send(ctx context.Context, convID, userID uuid.UUID, input message.SendInput) (*message.Message, error)
}

// ReactionAdder adds reactions for GraphQL mutations.
type ReactionAdder interface {
	Add(ctx context.Context, messageID, userID uuid.UUID, emoji string) error
}

// Handler is a minimal GraphQL-style JSON endpoint (prototype).
type Handler struct {
	conversations ConversationRepo
	messages      MessageSender
	reactions     ReactionAdder
	graphiql      bool
}

// NewHandler creates a GraphQL prototype handler.
func NewHandler(conv ConversationRepo, graphiql bool) *Handler {
	return &Handler{conversations: conv, graphiql: graphiql}
}

// SetMessageSender enables sendMessage mutation.
func (h *Handler) SetMessageSender(sender MessageSender) {
	h.messages = sender
}

// SetReactionAdder enables addReaction mutation.
func (h *Handler) SetReactionAdder(adder ReactionAdder) {
	h.reactions = adder
}

type gqlRequest struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables"`
	OperationName string         `json:"operationName"`
}

// HandleGraphQL serves POST /graphql with queries and mutations.
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
	case strings.Contains(q, "mutation") && strings.Contains(q, "addreaction"):
		h.handleAddReaction(w, r, claims.UserID, req.Variables)
	case strings.Contains(q, "mutation") && strings.Contains(q, "sendmessage"):
		h.handleSendMessage(w, r, claims.UserID, req.Variables)
	case strings.Contains(q, "conversations"):
		h.handleConversations(w, r, claims.UserID)
	default:
		apierror.WriteJSON(w, http.StatusOK, map[string]any{
			"data":   nil,
			"errors": []map[string]string{{"message": "unsupported; try conversations or mutation sendMessage"}},
		})
	}
}

func (h *Handler) handleSendMessage(w http.ResponseWriter, r *http.Request, userID uuid.UUID, vars map[string]any) {
	if h.messages == nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "mutations not configured")
		return
	}
	convRaw, _ := vars["conversationId"].(string)
	body, _ := vars["body"].(string)
	convID, err := uuid.Parse(convRaw)
	if err != nil || body == "" {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "conversationId and body required in variables")
		return
	}
	msg, err := h.messages.Send(r.Context(), convID, userID, message.SendInput{
		ClientMsgID: uuid.New().String(),
		Type:        message.TypeText,
		Body:        body,
	})
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"sendMessage": map[string]any{
				"id":   msg.ID,
				"seq":  msg.Seq,
				"body": msg.Body,
			},
		},
	})
}

func (h *Handler) handleAddReaction(w http.ResponseWriter, r *http.Request, userID uuid.UUID, vars map[string]any) {
	if h.reactions == nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "reactions not configured")
		return
	}
	msgRaw, _ := vars["messageId"].(string)
	emoji, _ := vars["emoji"].(string)
	msgID, err := uuid.Parse(msgRaw)
	if err != nil || emoji == "" {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "messageId and emoji required")
		return
	}
	if err := h.reactions.Add(r.Context(), msgID, userID, emoji); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{"addReaction": map[string]any{"message_id": msgID, "emoji": emoji}},
	})
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
<p>Query: <code>{"query":"{ conversations { id title } }"}</code></p>
<p>Mutation: <code>{"query":"mutation($conversationId:ID!,$body:String!){sendMessage(conversationId:$conversationId,body:$body){id seq body}}","variables":{"conversationId":"...","body":"hi"}}</code></p>
</body></html>`
