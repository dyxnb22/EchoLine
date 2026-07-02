package device

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

// ListRepository lists devices for a user.
type ListRepository struct {
	pool *pgxpool.Pool
}

// NewListRepository creates a list repository.
func NewListRepository(pool *pgxpool.Pool) *ListRepository {
	return &ListRepository{pool: pool}
}

// ListForUser returns all devices for a user ordered by last_seen_at DESC.
func (r *ListRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]Device, error) {
	const q = `
		SELECT id, user_id, device_name, platform, last_seen_at, created_at
		FROM devices
		WHERE user_id = $1
		ORDER BY last_seen_at DESC NULLS LAST
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.ID, &d.UserID, &d.DeviceName, &d.Platform, &d.LastSeenAt, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// Handler exposes device REST endpoints.
type Handler struct {
	repo *ListRepository
}

// NewHandler creates a device handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{repo: NewListRepository(pool)}
}

// HandleList lists devices for the authenticated user.
// GET /api/devices
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	devices, err := h.repo.ListForUser(r.Context(), claims.UserID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list devices")
		return
	}

	type item struct {
		ID         uuid.UUID  `json:"id"`
		DeviceName string     `json:"device_name"`
		Platform   string     `json:"platform"`
		LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
		CreatedAt  time.Time  `json:"created_at"`
	}
	out := make([]item, 0, len(devices))
	for _, d := range devices {
		out = append(out, item{
			ID:         d.ID,
			DeviceName: d.DeviceName,
			Platform:   d.Platform,
			LastSeenAt: d.LastSeenAt,
			CreatedAt:  d.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"devices": out})
}
