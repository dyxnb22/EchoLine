package reaction

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// Handler exposes reaction REST endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a reaction handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// HandleAdd adds a reaction.
// POST /api/messages/{message_id}/reactions
func (h *Handler) HandleAdd(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	msgID, err := parseMessageID(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid message_id")
		return
	}

	var req struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Emoji == "" {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "emoji is required")
		return
	}

	rx, err := h.repo.Add(r.Context(), msgID, claims.UserID, req.Emoji)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to add reaction")
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, map[string]any{
		"message_id": rx.MessageID,
		"user_id":    rx.UserID,
		"emoji":      rx.Emoji,
		"created_at": rx.CreatedAt.UTC().Format(time.RFC3339),
	})
}

// HandleRemove removes a reaction.
// DELETE /api/messages/{message_id}/reactions/{emoji}
func (h *Handler) HandleRemove(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	msgID, emoji, err := parseMessageIDAndEmoji(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid path")
		return
	}

	if err := h.repo.Remove(r.Context(), msgID, claims.UserID, emoji); err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to remove reaction")
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// HandleList lists reactions for a message.
// GET /api/messages/{message_id}/reactions
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	msgID, err := parseMessageID(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid message_id")
		return
	}

	reactions, err := h.repo.ListByMessage(r.Context(), msgID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list reactions")
		return
	}

	items := make([]map[string]any, 0, len(reactions))
	for _, rx := range reactions {
		items = append(items, map[string]any{
			"message_id": rx.MessageID,
			"user_id":    rx.UserID,
			"emoji":      rx.Emoji,
			"created_at": rx.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{"reactions": items})
}

// parseMessageID extracts message_id from /api/messages/{message_id}/reactions
func parseMessageID(path string) (uuid.UUID, error) {
	const prefix = "/api/messages/"
	if !strings.HasPrefix(path, prefix) {
		return uuid.Nil, errInvalidPath
	}
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.SplitN(rest, "/", 2)
	return uuid.Parse(parts[0])
}

// parseMessageIDAndEmoji extracts message_id and emoji from /api/messages/{message_id}/reactions/{emoji}
func parseMessageIDAndEmoji(path string) (uuid.UUID, string, error) {
	const prefix = "/api/messages/"
	if !strings.HasPrefix(path, prefix) {
		return uuid.Nil, "", errInvalidPath
	}
	rest := strings.TrimPrefix(path, prefix)
	// rest = {message_id}/reactions/{emoji}
	parts := strings.Split(rest, "/")
	if len(parts) < 3 || parts[1] != "reactions" {
		return uuid.Nil, "", errInvalidPath
	}
	msgID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, "", err
	}
	return msgID, parts[2], nil
}

var errInvalidPath = &pathError{}

type pathError struct{}

func (e *pathError) Error() string { return "invalid path" }
