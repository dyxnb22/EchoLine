package thread

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/message"
)

// Handler exposes thread REST endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a thread handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
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

	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Body == "" {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "body is required")
		return
	}

	msg, err := h.repo.SendReply(r.Context(), convID, claims.UserID, parentMsgID, req.Body)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to send reply")
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, msgPayload(msg))
}

// HandleListReplies lists replies for a parent message.
// GET /api/conversations/{conv_id}/messages/{message_id}/replies
func (h *Handler) HandleListReplies(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	_, parentMsgID, err := parseThreadPath(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid path")
		return
	}

	replies, err := h.repo.ListReplies(r.Context(), parentMsgID)
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
	// rest = {conv_id}/messages/{msg_id}/replies
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
	return map[string]any{
		"id":              msg.ID,
		"conversation_id": msg.ConversationID,
		"sender_id":       msg.SenderID,
		"client_msg_id":   msg.ClientMsgID,
		"seq":             msg.Seq,
		"type":            msg.Type,
		"body":            msg.Body,
		"status":          msg.Status,
		"created_at":      msg.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":      msg.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

var errInvalidPath = errPath("invalid path")

type errPath string

func (e errPath) Error() string { return string(e) }
