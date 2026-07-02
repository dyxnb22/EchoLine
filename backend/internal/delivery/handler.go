package delivery

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/message"
)

// Handler exposes delivery ACK REST endpoints.
type Handler struct {
	repo          *Repository
	conversations *conversation.Repository
	messages      *message.Repository
}

// NewHandler creates a delivery handler.
func NewHandler(repo *Repository, conversations *conversation.Repository, messages *message.Repository) *Handler {
	return &Handler{repo: repo, conversations: conversations, messages: messages}
}

type ackRequest struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Seq            int64  `json:"seq"`
	Status         string `json:"status"`
	DeviceID       string `json:"device_id"`
}

// HandleACK records delivered/read state for a message.
func (h *Handler) HandleACK(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req ackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	convID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}
	msgID, err := uuid.Parse(req.MessageID)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid message_id")
		return
	}

	member, err := h.conversations.IsMember(r.Context(), convID, claims.UserID)
	if err != nil || !member {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
		return
	}

	status := Status(req.Status)
	if status != StatusDelivered && status != StatusRead {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "status must be delivered or read")
		return
	}

	var readSeq int64
	if h.messages != nil {
		msg, err := h.messages.GetByID(r.Context(), convID, msgID)
		if err != nil {
			apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "message not in conversation")
			return
		}
		readSeq = msg.Seq
	}

	rec, err := h.repo.UpsertACK(r.Context(), msgID, claims.UserID, req.DeviceID, status)
	if err != nil {
		if errors.Is(err, ErrInvalidTransition) {
			apierror.Write(w, r, http.StatusConflict, "invalid_transition", "status cannot move backward")
			return
		}
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to record ack")
		return
	}

	if status == StatusRead && readSeq > 0 {
		_ = h.conversations.MarkRead(r.Context(), convID, claims.UserID, readSeq)
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"message_id": rec.MessageID,
		"status":     rec.Status,
		"acked_at":   rec.AckedAt,
	})
}
