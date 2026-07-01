package sync

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/message"
)

// Handler exposes offline sync endpoints.
type Handler struct {
	conversations *conversation.Repository
	messages      *message.Service
}

// NewHandler creates a sync handler.
func NewHandler(conversations *conversation.Repository, messages *message.Service) *Handler {
	return &Handler{conversations: conversations, messages: messages}
}

type cursor struct {
	ConversationID string `json:"conversation_id"`
	LastSeq        int64  `json:"last_seq"`
}

type syncRequest struct {
	DeviceID string   `json:"device_id"`
	Cursors  []cursor `json:"cursors"`
}

// HandleSync returns messages newer than provided cursors.
func (h *Handler) HandleSync(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req syncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	type convSync struct {
		ConversationID string           `json:"conversation_id"`
		Messages       []map[string]any `json:"messages"`
		LatestSeq      int64            `json:"latest_seq"`
		Unread         int64            `json:"unread"`
	}

	results := make([]convSync, 0, len(req.Cursors))
	for _, c := range req.Cursors {
		convID, err := uuid.Parse(c.ConversationID)
		if err != nil {
			apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
			return
		}

		member, err := h.conversations.IsMember(r.Context(), convID, claims.UserID)
		if err != nil || !member {
			apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
			return
		}

		msgs, err := h.messages.ListSince(r.Context(), convID, c.LastSeq, 200)
		if err != nil {
			apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to sync messages")
			return
		}

		state, err := h.conversations.GetMemberState(r.Context(), convID, claims.UserID)
		if err != nil {
			apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to load read state")
			return
		}

		items := make([]map[string]any, 0, len(msgs))
		for i := range msgs {
			items = append(items, message.ToCreatedPayload(&msgs[i]))
		}

		unread := state.LatestSeq - state.LastReadSeq
		if unread < 0 {
			unread = 0
		}

		results = append(results, convSync{
			ConversationID: convID.String(),
			Messages:       items,
			LatestSeq:      state.LatestSeq,
			Unread:         unread,
		})
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"device_id":     req.DeviceID,
		"conversations": results,
	})
}
