package message

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/audit"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/media"
	"github.com/echoline/echoline/backend/internal/metrics"
	"github.com/echoline/echoline/backend/internal/validate"
)

// Handler exposes message REST endpoints.
type Handler struct {
	service       *Service
	conversations *conversation.Repository
	attachments   *media.Repository
	audit         *audit.Repository
	webhook       WebhookNotifier
}

// NewHandler creates a message handler.
func NewHandler(service *Service, conversations *conversation.Repository, attachments *media.Repository, auditRepo *audit.Repository) *Handler {
	return &Handler{
		service:       service,
		conversations: conversations,
		attachments:   attachments,
		audit:         auditRepo,
		webhook:       noopWebhook{},
	}
}

type sendRequest struct {
	ClientMsgID string `json:"client_msg_id"`
	Type        string `json:"type"`
	Body        string `json:"body"`
	Attachment  *struct {
		ObjectKey string `json:"object_key"`
	} `json:"attachment"`
}

// HandleSend creates a message in a conversation.
func (h *Handler) HandleSend(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer metrics.ObserveMessageSend(start)

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

	input := SendInput{
		ClientMsgID: req.ClientMsgID,
		Type:        Type(req.Type),
		Body:        req.Body,
	}
	if req.Attachment != nil {
		input.ObjectKey = req.Attachment.ObjectKey
	}

	msg, err := h.service.Send(r.Context(), convID, claims.UserID, input)
	if err != nil {
		if errors.Is(err, conversation.ErrNotMember) {
			apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
			return
		}
		if errors.Is(err, conversation.ErrCannotPublish) {
			apierror.Write(w, r, http.StatusForbidden, "forbidden", "cannot publish to this conversation")
			return
		}
		if errors.Is(err, media.ErrAttachmentNotFound) {
			apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "attachment not found")
			return
		}
		if errors.Is(err, ErrBlocked) {
			apierror.Write(w, r, http.StatusForbidden, "blocked", "recipient has blocked you")
			return
		}
		if errors.Is(err, validate.ErrMessageBodyEmpty) || errors.Is(err, validate.ErrMessageBodyLong) {
			apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, ToCreatedPayload(msg))
	h.notifyWebhook(r.Context(), msg)
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

	attachmentByMsg := map[uuid.UUID]media.Attachment{}
	if h.attachments != nil && len(messages) > 0 {
		ids := make([]uuid.UUID, 0, len(messages))
		for i := range messages {
			ids = append(ids, messages[i].ID)
		}
		if m, err := h.attachments.ListByMessageIDs(r.Context(), ids); err == nil {
			attachmentByMsg = m
		}
	}

	items := make([]map[string]any, 0, len(messages))
	for i := range messages {
		var att *media.Attachment
		if a, ok := attachmentByMsg[messages[i].ID]; ok {
			att = &a
		}
		items = append(items, ToCreatedPayloadWithAttachment(&messages[i], att))
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

// HandleEdit updates a message body.
func (h *Handler) HandleEdit(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, msgID, err := parseMessagePath(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid path")
		return
	}

	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	msg, err := h.service.Edit(r.Context(), convID, msgID, claims.UserID, req.Body)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apierror.Write(w, r, http.StatusNotFound, "not_found", "message not found")
			return
		}
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusOK, ToCreatedPayload(msg))
}

// HandleRecall marks a message as recalled.
func (h *Handler) HandleRecall(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, msgID, err := parseMessagePath(strings.TrimSuffix(r.URL.Path, "/recall"))
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid path")
		return
	}

	msg, err := h.service.Recall(r.Context(), convID, msgID, claims.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apierror.Write(w, r, http.StatusNotFound, "not_found", "message not found")
			return
		}
		if errors.Is(err, conversation.ErrForbidden) {
			apierror.Write(w, r, http.StatusForbidden, "forbidden", "cannot recall message")
			return
		}
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if h.audit != nil {
		_ = h.audit.LogRecall(r.Context(), claims.UserID, msg.ID.String(), convID.String(), msg.Seq)
	}

	apierror.WriteJSON(w, http.StatusOK, ToCreatedPayload(msg))
}

func parseMessagePath(path string) (uuid.UUID, uuid.UUID, error) {
	const prefix = "/api/conversations/"
	if !strings.HasPrefix(path, prefix) {
		return uuid.Nil, uuid.Nil, errors.New("invalid path")
	}
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.Split(rest, "/")
	if len(parts) < 3 || parts[1] != "messages" {
		return uuid.Nil, uuid.Nil, errors.New("invalid path")
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
