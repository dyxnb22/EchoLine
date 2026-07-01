package recommendation

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// ChannelSummary is a recommended channel.
type ChannelSummary struct {
	ID        uuid.UUID
	Title     string
	CreatedAt time.Time
}

// Repository fetches channel recommendations.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a recommendation repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// RecommendChannels returns channels the user has not joined, ranked by subscriber count.
func (r *Repository) RecommendChannels(ctx context.Context, userID uuid.UUID, limit int) ([]ChannelSummary, error) {
	const q = `
		SELECT c.id, c.title, c.created_at
		FROM conversations c
		LEFT JOIN conversation_members cm ON cm.conversation_id = c.id
		WHERE c.type = 'channel'
		  AND c.id NOT IN (
			SELECT conversation_id FROM conversation_members WHERE user_id = $1
		  )
		GROUP BY c.id, c.title, c.created_at
		ORDER BY COUNT(cm.user_id) DESC, c.latest_seq DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("recommend channels: %w", err)
	}
	defer rows.Close()

	var out []ChannelSummary
	for rows.Next() {
		var cs ChannelSummary
		if err := rows.Scan(&cs.ID, &cs.Title, &cs.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		out = append(out, cs)
	}
	return out, rows.Err()
}

// UserSummary is a recommended contact.
type UserSummary struct {
	ID          uuid.UUID
	Username    string
	DisplayName string
}

// RecommendFriends returns users who share a group with the requester but are not blocked.
func (r *Repository) RecommendFriends(ctx context.Context, userID uuid.UUID, limit int) ([]UserSummary, error) {
	const q = `
		SELECT DISTINCT u.id, u.username, u.display_name
		FROM users u
		JOIN conversation_members cm1 ON cm1.user_id = u.id
		JOIN conversation_members cm2 ON cm2.conversation_id = cm1.conversation_id
		JOIN conversations c ON c.id = cm1.conversation_id AND c.type IN ('group', 'direct')
		WHERE cm2.user_id = $1
		  AND u.id != $1
		  AND u.id NOT IN (SELECT blocked_id FROM user_blocks WHERE blocker_id = $1)
		ORDER BY u.username
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("recommend friends: %w", err)
	}
	defer rows.Close()

	var out []UserSummary
	for rows.Next() {
		var u UserSummary
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName); err != nil {
			return nil, fmt.Errorf("scan friend: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// Handler exposes recommendation REST endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a recommendation handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// HandleRecommendChannels returns recommended channels for the user.
// GET /api/recommendations/channels
func (h *Handler) HandleRecommendChannels(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	channels, err := h.repo.RecommendChannels(r.Context(), claims.UserID, 20)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to fetch recommendations")
		return
	}

	items := make([]map[string]any, 0, len(channels))
	for _, ch := range channels {
		items = append(items, map[string]any{
			"id":         ch.ID,
			"title":      ch.Title,
			"created_at": ch.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{"channels": items})
}

// HandleRecommendFriends returns friend suggestions based on mutual groups.
// GET /api/recommendations/friends
func (h *Handler) HandleRecommendFriends(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	friends, err := h.repo.RecommendFriends(r.Context(), claims.UserID, 20)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to fetch friends")
		return
	}

	items := make([]map[string]any, 0, len(friends))
	for _, f := range friends {
		items = append(items, map[string]any{
			"id":           f.ID,
			"username":     f.Username,
			"display_name": f.DisplayName,
		})
	}
	apierror.WriteJSON(w, http.StatusOK, map[string]any{"friends": items})
}
