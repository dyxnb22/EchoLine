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
	dlq    *DLQRepository
	outbox *Repository
}

// NewDLQReplayHandler creates a DLQ replay handler.
func NewDLQReplayHandler(dlq *DLQRepository, outbox *Repository) *DLQReplayHandler {
	return &DLQReplayHandler{dlq: dlq, outbox: outbox}
}

// HandleReplay re-queues a dead letter event into the outbox.
// POST /api/admin/dlq/{id}/replay
func (h *DLQReplayHandler) HandleReplay(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	path := strings.TrimSuffix(r.URL.Path, "/replay")
	parts := strings.Split(path, "/")
	rawID := parts[len(parts)-1]

	id, err := uuid.Parse(rawID)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid dlq event id")
		return
	}

	evt, err := h.dlq.GetByID(r.Context(), id)
	if err != nil {
		apierror.Write(w, r, http.StatusNotFound, "not_found", "dlq event not found")
		return
	}

	if err := h.outbox.Enqueue(r.Context(), evt.SourceTopic, evt.Payload); err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to requeue event")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":           evt.ID,
		"source_topic": evt.SourceTopic,
		"status":       "replay_queued",
	})
}
