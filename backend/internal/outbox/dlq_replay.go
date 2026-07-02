package outbox

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// DLQReplayHandler exposes admin DLQ replay endpoint.
type DLQReplayHandler struct {
	repo *DLQRepository
}

// NewDLQReplayHandler creates a DLQ replay handler.
func NewDLQReplayHandler(repo *DLQRepository) *DLQReplayHandler {
	return &DLQReplayHandler{repo: repo}
}

// HandleReplay re-queues a dead letter event by ID (skeleton).
// POST /api/admin/dlq/{id}/replay
func (h *DLQReplayHandler) HandleReplay(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	// Extract id from path /api/admin/dlq/{id}/replay
	path := strings.TrimSuffix(r.URL.Path, "/replay")
	parts := strings.Split(path, "/")
	rawID := parts[len(parts)-1]

	id, err := uuid.Parse(rawID)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid dlq event id")
		return
	}

	events, err := h.repo.ListLast50(r.Context())
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to access DLQ")
		return
	}

	var found *DeadLetterEvent
	for i := range events {
		if events[i].ID == id {
			found = &events[i]
			break
		}
	}
	if found == nil {
		apierror.Write(w, r, http.StatusNotFound, "not_found", "dlq event not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":           found.ID,
		"source_topic": found.SourceTopic,
		"status":       "replay_queued",
	})
}
