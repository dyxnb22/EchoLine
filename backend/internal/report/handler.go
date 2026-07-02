package report

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

// Handler exposes the report REST endpoint.
type Handler struct {
	repo     *Repository
	convRepo *conversation.Repository
	messages *message.Repository
}

// NewHandler creates a report handler.
func NewHandler(repo *Repository, convRepo *conversation.Repository, messages *message.Repository) *Handler {
	return &Handler{repo: repo, convRepo: convRepo, messages: messages}
}

type createReportRequest struct {
	Reason string `json:"reason"`
}

// HandleCreate creates a report for a message.
func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	// path: /api/conversations/{id}/messages/{message_id}/report
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid path")
		return
	}

	convID, err := uuid.Parse(parts[2])
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	msgID, err := uuid.Parse(parts[4])
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid message_id")
		return
	}

	member, err := h.convRepo.IsMember(r.Context(), convID, claims.UserID)
	if err != nil || !member {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
		return
	}

	if _, err := h.messages.GetByID(r.Context(), convID, msgID); err != nil {
		if errors.Is(err, message.ErrNotFound) {
			apierror.Write(w, r, http.StatusNotFound, "not_found", "message not found")
			return
		}
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to resolve message")
		return
	}

	var req createReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}

	rpt, err := h.repo.Create(r.Context(), claims.UserID, msgID, convID, req.Reason)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to create report")
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, map[string]any{
		"id":              rpt.ID,
		"message_id":      rpt.MessageID,
		"conversation_id": rpt.ConversationID,
		"reason":          rpt.Reason,
		"created_at":      rpt.CreatedAt,
	})
}
