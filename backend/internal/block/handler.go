package block

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// Handler exposes block/unblock REST endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a block handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func parseBlockedID(path string) (uuid.UUID, error) {
	// path: /api/blocks/{user_id}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return uuid.Nil, errors.New("invalid path")
	}
	return uuid.Parse(parts[2])
}

// HandleBlock blocks a user.
func (h *Handler) HandleBlock(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	blockedID, err := parseBlockedID(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid user_id")
		return
	}

	b, err := h.repo.Block(r.Context(), claims.UserID, blockedID)
	if err != nil {
		if errors.Is(err, ErrSelfBlock) {
			apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "cannot block yourself")
			return
		}
		if errors.Is(err, ErrAlreadyBlocked) {
			apierror.Write(w, r, http.StatusConflict, "already_blocked", "user already blocked")
			return
		}
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to block")
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, map[string]any{
		"blocker_id": b.BlockerID,
		"blocked_id": b.BlockedID,
		"created_at": b.CreatedAt,
	})
}

// HandleUnblock unblocks a user.
func (h *Handler) HandleUnblock(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	blockedID, err := parseBlockedID(r.URL.Path)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid user_id")
		return
	}

	if err := h.repo.Unblock(r.Context(), claims.UserID, blockedID); err != nil {
		if errors.Is(err, ErrNotBlocked) {
			apierror.Write(w, r, http.StatusNotFound, "not_found", "block not found")
			return
		}
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to unblock")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleList lists users blocked by the caller.
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	blocks, err := h.repo.List(r.Context(), claims.UserID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list blocks")
		return
	}

	type item struct {
		BlockedID uuid.UUID `json:"blocked_id"`
		CreatedAt string    `json:"created_at"`
	}
	out := make([]item, 0, len(blocks))
	for _, b := range blocks {
		out = append(out, item{
			BlockedID: b.BlockedID,
			CreatedAt: b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"blocks": out})
}
