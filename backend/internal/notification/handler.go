package notification

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// Handler exposes notification REST endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a notification handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// HandleList returns notifications for the authenticated user.
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	events, err := h.repo.ListForUser(r.Context(), claims.UserID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list notifications")
		return
	}

	type item struct {
		ID        uuid.UUID       `json:"id"`
		Type      string          `json:"type"`
		Payload   json.RawMessage `json:"payload"`
		ReadAt    *string         `json:"read_at,omitempty"`
		CreatedAt string          `json:"created_at"`
	}
	out := make([]item, 0, len(events))
	for _, e := range events {
		it := item{
			ID:        e.ID,
			Type:      e.Type,
			Payload:   e.Payload,
			CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if e.ReadAt != nil {
			s := e.ReadAt.Format("2006-01-02T15:04:05Z07:00")
			it.ReadAt = &s
		}
		out = append(out, it)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"notifications": out})
}

// HandleMarkRead marks a single notification as read.
// Path: /api/notifications/{id}/read
func (h *Handler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid path")
		return
	}
	id, err := uuid.Parse(parts[2])
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid notification_id")
		return
	}

	if err := h.repo.MarkRead(r.Context(), id, claims.UserID); err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to mark read")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleMarkAllRead marks all notifications as read.
func (h *Handler) HandleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	if err := h.repo.MarkAllRead(r.Context(), claims.UserID); err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to mark all read")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
