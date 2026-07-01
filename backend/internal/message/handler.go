package message

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/conversation"
)

// Handler exposes message REST endpoints.
type Handler struct {
	messages      *Repository
	conversations *conversation.Repository
}

// NewHandler creates a message handler.
func NewHandler(messages *Repository, conversations *conversation.Repository) *Handler {
	return &Handler{
		messages:      messages,
		conversations: conversations,
	}
}

type sendRequest struct {
	ClientMsgID string `json:"client_msg_id"`
	Type        string `json:"type"`
	Body        string `json:"body"`
}

// HandleSend creates a message in a conversation.
func (h *Handler) HandleSend(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, err := conversation.ParseConversationID(r.URL.Path, "/api/conversations/", "/messages")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	member, err := h.conversations.IsMember(r.Context(), convID, claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to check membership")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "forbidden", "not a conversation member")
		return
	}

	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	msg, err := h.messages.Create(r.Context(), convID, claims.UserID, req.ClientMsgID, Type(req.Type), req.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toMessageResponse(msg))
}

// HandleList returns paginated conversation messages.
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	path := strings.TrimSuffix(r.URL.Path, "/")
	convID, err := conversation.ParseConversationID(path, "/api/conversations/", "/messages")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	member, err := h.conversations.IsMember(r.Context(), convID, claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to check membership")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "forbidden", "not a conversation member")
		return
	}

	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	var beforeSeq *int64
	if raw := r.URL.Query().Get("before_seq"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid before_seq")
			return
		}
		beforeSeq = &parsed
	}

	messages, err := h.messages.List(r.Context(), convID, beforeSeq, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list messages")
		return
	}

	items := make([]map[string]any, 0, len(messages))
	for _, msg := range messages {
		items = append(items, toMessageResponse(&msg))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"messages": items,
	})
}

func toMessageResponse(msg *Message) map[string]any {
	return map[string]any{
		"id":              msg.ID,
		"conversation_id": msg.ConversationID,
		"sender_id":       msg.SenderID,
		"client_msg_id":   msg.ClientMsgID,
		"seq":             msg.Seq,
		"type":            msg.Type,
		"body":            msg.Body,
		"status":          msg.Status,
		"created_at":      msg.CreatedAt,
		"updated_at":      msg.UpdatedAt,
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
