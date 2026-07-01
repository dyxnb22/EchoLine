package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// MuteRepository manages mute state for conversation members.
type MuteRepository struct {
	pool *pgxpool.Pool
}

// NewMuteRepository creates a mute repository.
func NewMuteRepository(pool *pgxpool.Pool) *MuteRepository {
	return &MuteRepository{pool: pool}
}

// Mute sets muted_until for the user in the conversation.
// duration == 0 means mute indefinitely (year 9999).
func (r *MuteRepository) Mute(ctx context.Context, convID, userID uuid.UUID, duration time.Duration) error {
	until := time.Now().UTC().Add(duration)
	if duration == 0 {
		until = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}
	const q = `
		UPDATE conversation_members
		SET muted_until = $1
		WHERE conversation_id = $2 AND user_id = $3
	`
	tag, err := r.pool.Exec(ctx, q, until, convID, userID)
	if err != nil {
		return fmt.Errorf("mute conversation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotMember
	}
	return nil
}

// Unmute clears muted_until for the user in the conversation.
func (r *MuteRepository) Unmute(ctx context.Context, convID, userID uuid.UUID) error {
	const q = `
		UPDATE conversation_members
		SET muted_until = NULL
		WHERE conversation_id = $1 AND user_id = $2
	`
	tag, err := r.pool.Exec(ctx, q, convID, userID)
	if err != nil {
		return fmt.Errorf("unmute conversation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotMember
	}
	return nil
}

// MuteHandler handles mute/unmute REST endpoints.
type MuteHandler struct {
	repo    *MuteRepository
	convRep *Repository
}

// NewMuteHandler creates a mute handler.
func NewMuteHandler(pool *pgxpool.Pool, convRepo *Repository) *MuteHandler {
	return &MuteHandler{repo: NewMuteRepository(pool), convRep: convRepo}
}

type muteRequest struct {
	// DurationSeconds: 0 means indefinite. Omit or 0 for permanent mute.
	DurationSeconds int `json:"duration_seconds"`
}

// HandleMute mutes a conversation for the authenticated user.
// POST /api/conversations/{id}/mute
func (h *MuteHandler) HandleMute(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, err := ParseConversationID(r.URL.Path, "/api/conversations/", "/mute")
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	member, err := h.convRep.IsMember(r.Context(), convID, claims.UserID)
	if err != nil || !member {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
		return
	}

	var req muteRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	var dur time.Duration
	if req.DurationSeconds > 0 {
		dur = time.Duration(req.DurationSeconds) * time.Second
	}

	if err := h.repo.Mute(r.Context(), convID, claims.UserID, dur); err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to mute")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleUnmute unmutes a conversation for the authenticated user.
// POST /api/conversations/{id}/unmute
func (h *MuteHandler) HandleUnmute(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, err := ParseConversationID(r.URL.Path, "/api/conversations/", "/unmute")
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	member, err := h.convRep.IsMember(r.Context(), convID, claims.UserID)
	if err != nil || !member {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
		return
	}

	if err := h.repo.Unmute(r.Context(), convID, claims.UserID); err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to unmute")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
