package pin

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

// Handler exposes pin/unpin REST endpoints.
type Handler struct {
	repo     *Repository
	convRepo *conversation.Repository
	messages *message.Repository
}

// NewHandler creates a pin handler.
func NewHandler(repo *Repository, convRepo *conversation.Repository, messages *message.Repository) *Handler {
	return &Handler{repo: repo, convRepo: convRepo, messages: messages}
}

func parseConvAndMessageID(path string) (uuid.UUID, uuid.UUID, error) {
	// path: /api/conversations/{id}/pins/{message_id}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 5 {
		return uuid.Nil, uuid.Nil, errors.New("invalid path")
	}
	convID, err := uuid.Parse(parts[2])
	if err != nil {
		return uuid.Nil, uuid.Nil, errors.New("invalid conversation_id")
	}
	msgID, err := uuid.Parse(parts[4])
	if err != nil {
		return uuid.Nil, uuid.Nil, errors.New("invalid message_id")
	}
	return convID, msgID, nil
}

func parseConvID(path string) (uuid.UUID, error) {
	// path: /api/conversations/{id}/pins
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return uuid.Nil, errors.New("invalid path")
	}
	return uuid.Parse(parts[2])
}

func (h *Handler) requireMessageInConversation(w http.ResponseWriter, r *http.Request, convID, msgID uuid.UUID) bool {
	_, err := h.messages.GetByID(r.Context(), convID, msgID)
	if err != nil {
		if errors.Is(err, message.ErrNotFound) {
			apierror.Write(w, r, http.StatusNotFound, "not_found", "message not found")
			return false
		}
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to resolve message")
		return false
	}
	return true
}

// HandlePin pins a message.
func (h *Handler) HandlePin(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, msgID, err := parseConvAndMessageID(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	member, err := h.convRepo.IsMember(r.Context(), convID, claims.UserID)
	if err != nil || !member {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
		return
	}
	if !h.requireMessageInConversation(w, r, convID, msgID) {
		return
	}

	p, err := h.repo.Pin(r.Context(), convID, msgID, claims.UserID)
	if err != nil {
		if errors.Is(err, ErrAlreadyPinned) {
			apierror.Write(w, r, http.StatusConflict, "already_pinned", "message already pinned")
			return
		}
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to pin")
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, map[string]any{
		"conversation_id": p.ConversationID,
		"message_id":      p.MessageID,
		"pinned_by":       p.PinnedBy,
		"pinned_at":       p.PinnedAt,
	})
}

// HandleUnpin unpins a message.
func (h *Handler) HandleUnpin(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, msgID, err := parseConvAndMessageID(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	member, err := h.convRepo.IsMember(r.Context(), convID, claims.UserID)
	if err != nil || !member {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
		return
	}
	if !h.requireMessageInConversation(w, r, convID, msgID) {
		return
	}

	if err := h.repo.Unpin(r.Context(), convID, msgID); err != nil {
		if errors.Is(err, ErrNotPinned) {
			apierror.Write(w, r, http.StatusNotFound, "not_found", "pin not found")
			return
		}
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to unpin")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleList lists pinned messages for a conversation.
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, err := parseConvID(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	member, err := h.convRepo.IsMember(r.Context(), convID, claims.UserID)
	if err != nil || !member {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
		return
	}

	pins, err := h.repo.List(r.Context(), convID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list pins")
		return
	}

	type item struct {
		ConversationID uuid.UUID `json:"conversation_id"`
		MessageID      uuid.UUID `json:"message_id"`
		PinnedBy       uuid.UUID `json:"pinned_by"`
		PinnedAt       string    `json:"pinned_at"`
	}

	out := make([]item, 0, len(pins))
	for _, p := range pins {
		out = append(out, item{
			ConversationID: p.ConversationID,
			MessageID:      p.MessageID,
			PinnedBy:       p.PinnedBy,
			PinnedAt:       p.PinnedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"pins": out})
}
