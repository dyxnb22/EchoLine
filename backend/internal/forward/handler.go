package forward

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

// Handler exposes forward REST endpoints.
type Handler struct {
	messages *message.Service
}

// NewHandler creates a forward handler.
func NewHandler(messages *message.Service) *Handler {
	return &Handler{messages: messages}
}

// HandleForward forwards a message to another conversation.
// POST /api/messages/{message_id}/forward
func (h *Handler) HandleForward(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	msgIDStr := r.PathValue("message_id")
	if msgIDStr == "" {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "message_id required")
		return
	}
	sourceMsgID, err := uuid.Parse(msgIDStr)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid message_id")
		return
	}

	var req struct {
		TargetConversationID string `json:"target_conversation_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	targetConvID, err := uuid.Parse(req.TargetConversationID)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid target_conversation_id")
		return
	}

	msg, err := h.messages.Forward(r.Context(), sourceMsgID, targetConvID, claims.UserID)
	if err != nil {
		if errors.Is(err, message.ErrNotFound) {
			apierror.Write(w, r, http.StatusNotFound, "not_found", "source message not found")
			return
		}
		if errors.Is(err, conversation.ErrForbidden) || errors.Is(err, conversation.ErrCannotPublish) {
			apierror.Write(w, r, http.StatusForbidden, "forbidden", err.Error())
			return
		}
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, message.ToCreatedPayload(msg))
}
