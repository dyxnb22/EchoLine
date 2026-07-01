package message

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/conversation"
)

// Handler exposes message REST endpoints.
type Handler struct {
	service       *Service
	conversations *conversation.Repository
}

// NewHandler creates a message handler.
func NewHandler(service *Service, conversations *conversation.Repository) *Handler {
	return &Handler{
		service:       service,
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
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, err := conversation.ParseConversationID(r.URL.Path, "/api/conversations/", "/messages")
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	msg, err := h.service.Send(r.Context(), convID, claims.UserID, req.ClientMsgID, Type(req.Type), req.Body)
	if err != nil {
		if errors.Is(err, conversation.ErrNotMember) {
			apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
			return
		}
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, ToCreatedPayload(msg))
}

// HandleList returns paginated conversation messages.
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	path := strings.TrimSuffix(r.URL.Path, "/")
	convID, err := conversation.ParseConversationID(path, "/api/conversations/", "/messages")
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	member, err := h.conversations.IsMember(r.Context(), convID, claims.UserID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to check membership")
		return
	}
	if !member {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
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
			apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid before_seq")
			return
		}
		beforeSeq = &parsed
	}

	messages, err := h.service.List(r.Context(), convID, beforeSeq, limit)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list messages")
		return
	}

	items := make([]map[string]any, 0, len(messages))
	for i := range messages {
		items = append(items, ToCreatedPayload(&messages[i]))
	}

	var nextBefore *int64
	if len(messages) == limit {
		seq := messages[len(messages)-1].Seq
		nextBefore = &seq
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"messages":    items,
		"next_before": nextBefore,
	})
}

// HandleMarkRead updates last_read_seq for a conversation.
func (h *Handler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, err := conversation.ParseConversationID(r.URL.Path, "/api/conversations/", "/read")
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	var req struct {
		LastReadSeq int64 `json:"last_read_seq"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if err := h.conversations.MarkRead(r.Context(), convID, claims.UserID, req.LastReadSeq); err != nil {
		if errors.Is(err, conversation.ErrNotMember) {
			apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
			return
		}
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to mark read")
		return
	}

	state, _ := h.conversations.GetMemberState(r.Context(), convID, claims.UserID)
	unread := int64(0)
	if state != nil {
		unread = state.LatestSeq - state.LastReadSeq
		if unread < 0 {
			unread = 0
		}
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"conversation_id": convID,
		"last_read_seq":   req.LastReadSeq,
		"unread":          unread,
	})
}
