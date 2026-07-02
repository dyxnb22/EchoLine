package presence

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/redisx"
)

// LastSeenStore tracks user last-seen timestamps.
type LastSeenStore struct {
	redis *redisx.Client
	ttl   time.Duration
}

// NewLastSeenStore creates a Redis-backed last-seen store.
func NewLastSeenStore(redis *redisx.Client) *LastSeenStore {
	return &LastSeenStore{redis: redis, ttl: 7 * 24 * time.Hour}
}

// Set records last-seen for a user.
func (s *LastSeenStore) Set(ctx context.Context, userID string) error {
	if s == nil || s.redis == nil {
		return nil
	}
	key := fmt.Sprintf("lastseen:%s", userID)
	return s.redis.Set(ctx, key, time.Now().UTC().Format(time.RFC3339), s.ttl)
}

// Get returns last-seen RFC3339 timestamp or empty string.
func (s *LastSeenStore) Get(ctx context.Context, userID string) (string, error) {
	if s == nil || s.redis == nil {
		return "", nil
	}
	key := fmt.Sprintf("lastseen:%s", userID)
	val, err := s.redis.Get(ctx, key)
	if err != nil {
		return "", nil
	}
	return val, nil
}

// LastSeenHandler exposes last-seen REST endpoints.
type LastSeenHandler struct {
	store    *LastSeenStore
	contacts contactGate
}

// NewLastSeenHandler creates a last-seen handler.
func NewLastSeenHandler(store *LastSeenStore) *LastSeenHandler {
	return &LastSeenHandler{store: store}
}

// SetContactGate restricts lookups to users who share a conversation with the caller.
func (h *LastSeenHandler) SetContactGate(g contactGate) {
	h.contacts = g
}

// HandleGet returns last-seen for user IDs.
// GET /api/presence/last-seen?user_ids=...
func (h *LastSeenHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	raw := r.URL.Query().Get("user_ids")
	result := make(map[string]string)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		targetID, err := uuid.Parse(part)
		if err != nil {
			continue
		}
		if h.contacts != nil && targetID != claims.UserID {
			shared, err := h.contacts.ShareAnyConversation(r.Context(), claims.UserID, targetID)
			if err != nil || !shared {
				continue
			}
		}
		if h.store != nil {
			if ts, _ := h.store.Get(r.Context(), part); ts != "" {
				result[part] = ts
			}
		}
	}
	apierror.WriteJSON(w, http.StatusOK, map[string]any{"last_seen": result})
}

// HandleTouch updates last-seen for the current user.
// POST /api/presence/last-seen
func (h *LastSeenHandler) HandleTouch(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}
	if h.store != nil {
		_ = h.store.Set(r.Context(), claims.UserID.String())
	}
	apierror.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
