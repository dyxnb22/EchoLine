package search

import (
	"net/http"
	"strconv"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// Handler exposes search endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a search handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// HandleSearch performs keyword search scoped to member conversations.
func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	q := r.URL.Query().Get("q")
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	hits, err := h.repo.Search(r.Context(), claims.UserID, q, limit)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	items := make([]map[string]any, 0, len(hits))
	for _, hit := range hits {
		items = append(items, map[string]any{
			"message_id":      hit.MessageID,
			"conversation_id": hit.ConversationID,
			"sender_id":       hit.SenderID,
			"body":            hit.Body,
			"seq":             hit.Seq,
			"created_at":      hit.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"query":   q,
		"results": items,
	})
}
