package thread

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/message"
)

// Handler exposes thread REST endpoints.
type Handler struct {
	messages      *message.Service
	conversations *conversation.Repository
	repo          *Repository
}

// NewHandler creates a thread handler.
func NewHandler(messages *message.Service, conversations *conversation.Repository, repo *Repository) *Handler {
	return &Handler{messages: messages, conversations: conversations, repo: repo}
}

// HandleSendReply sends a reply to a parent message.
// POST /api/conversations/{conv_id}/messages/{message_id}/replies
func (h *Handler) HandleSendReply(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, parentMsgID, err := parseThreadPath(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid path")
		return
	}

	member, err := h.conversations.IsMember(r.Context(), convID, claims.UserID)
	if err != nil || !member {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
		return
	}

	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Body == "" {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "body is required")
		return
	}

	msg, err := h.messages.SendReply(r.Context(), convID, claims.UserID, parentMsgID, req.Body)
	if err != nil {
		if errors.Is(err, conversation.ErrForbidden) || errors.Is(err, conversation.ErrCannotPublish) {
			apierror.Write(w, r, http.StatusForbidden, "forbidden", err.Error())
			return
		}
		if errors.Is(err, message.ErrNotFound) {
			apierror.Write(w, r, http.StatusNotFound, "not_found", "parent message not found")
			return
		}
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, msgPayload(msg))
}

// HandleListReplies lists replies for a parent message.
// GET /api/conversations/{conv_id}/messages/{message_id}/replies
func (h *Handler) HandleListReplies(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, parentMsgID, err := parseThreadPath(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid path")
		return
	}

	member, err := h.conversations.IsMember(r.Context(), convID, claims.UserID)
	if err != nil || !member {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
		return
	}

	if _, err := h.messages.GetByID(r.Context(), convID, parentMsgID); err != nil {
		if errors.Is(err, message.ErrNotFound) {
			apierror.Write(w, r, http.StatusNotFound, "not_found", "parent message not found")
			return
		}
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to resolve parent message")
		return
	}

	replies, err := h.repo.ListReplies(r.Context(), convID, parentMsgID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list replies")
		return
	}

	items := make([]map[string]any, 0, len(replies))
	for i := range replies {
		items = append(items, msgPayload(&replies[i]))
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{"replies": items})
}

// parseThreadPath extracts conv_id and message_id from
// /api/conversations/{conv_id}/messages/{message_id}/replies
func parseThreadPath(path string) (uuid.UUID, uuid.UUID, error) {
	const prefix = "/api/conversations/"
	if !strings.HasPrefix(path, prefix) {
		return uuid.Nil, uuid.Nil, errInvalidPath
	}
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.Split(rest, "/")
	if len(parts) < 4 || parts[1] != "messages" {
		return uuid.Nil, uuid.Nil, errInvalidPath
	}
	convID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	msgID, err := uuid.Parse(parts[2])
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return convID, msgID, nil
}

func msgPayload(msg *message.Message) map[string]any {
	return message.ToCreatedPayload(msg)
}

var errInvalidPath = errPath("invalid path")

type errPath string

func (e errPath) Error() string { return string(e) }
