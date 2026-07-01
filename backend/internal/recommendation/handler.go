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

// RecommendChannels returns channels the user has not joined, ordered by member count.
func (r *Repository) RecommendChannels(ctx context.Context, userID uuid.UUID, limit int) ([]ChannelSummary, error) {
	const q = `
		SELECT c.id, c.title, c.created_at
		FROM conversations c
		WHERE c.type = 'channel'
		  AND c.id NOT IN (
			SELECT conversation_id FROM conversation_members WHERE user_id = $1
		  )
		ORDER BY c.latest_seq DESC
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
