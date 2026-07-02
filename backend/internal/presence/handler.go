package presence

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/redisx"
)

// OnlineChecker checks whether user IDs are currently online.
type OnlineChecker interface {
	IsOnline(ctx context.Context, userID string) (bool, error)
}

// RedisOnlineChecker implements OnlineChecker via Redis presence keys.
type RedisOnlineChecker struct {
	redis *redisx.Client
}

// NewRedisOnlineChecker creates a checker backed by Redis.
func NewRedisOnlineChecker(redis *redisx.Client) *RedisOnlineChecker {
	return &RedisOnlineChecker{redis: redis}
}

// IsOnline returns true when any presence key exists for the user.
func (c *RedisOnlineChecker) IsOnline(ctx context.Context, userID string) (bool, error) {
	return c.redis.HasPresence(ctx, userID)
}

// OnlineHandler exposes GET /api/presence/online endpoint.
type OnlineHandler struct {
	checker OnlineChecker
}

// NewOnlineHandler creates a presence online handler.
func NewOnlineHandler(checker OnlineChecker) *OnlineHandler {
	return &OnlineHandler{checker: checker}
}

// HandleOnline returns online status for a list of user_ids.
// GET /api/presence/online?user_ids=uuid1,uuid2,...
func (h *OnlineHandler) HandleOnline(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	raw := r.URL.Query().Get("user_ids")
	if raw == "" {
		apierror.WriteJSON(w, http.StatusOK, map[string]any{"online": map[string]bool{}})
		return
	}

	parts := strings.Split(raw, ",")
	result := make(map[string]bool, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if _, err := uuid.Parse(part); err != nil {
			continue
		}
		if h.checker == nil {
			result[part] = false
			continue
		}
		online, err := h.checker.IsOnline(r.Context(), part)
		if err != nil {
			result[part] = false
		} else {
			result[part] = online
		}
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{"online": result})
}
