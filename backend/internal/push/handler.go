package push

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

// Token represents a push notification token for a device.
type Token struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	DeviceID  string
	Token     string
	Platform  string
	CreatedAt time.Time
}

// Repository persists push tokens.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a push token repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Register upserts a push token for a user/device.
func (r *Repository) Register(ctx context.Context, userID uuid.UUID, deviceID, token, platform string) (*Token, error) {
	const q = `
		INSERT INTO push_tokens (id, user_id, device_id, token, platform, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
		ON CONFLICT (user_id, device_id) DO UPDATE
			SET token = EXCLUDED.token, platform = EXCLUDED.platform, created_at = EXCLUDED.created_at
		RETURNING id, user_id, device_id, token, platform, created_at
	`
	now := time.Now().UTC()
	row := r.pool.QueryRow(ctx, q, userID, deviceID, token, platform, now)
	var t Token
	if err := row.Scan(&t.ID, &t.UserID, &t.DeviceID, &t.Token, &t.Platform, &t.CreatedAt); err != nil {
		return nil, fmt.Errorf("register push token: %w", err)
	}
	return &t, nil
}

// List returns all push tokens for a user.
func (r *Repository) List(ctx context.Context, userID uuid.UUID) ([]Token, error) {
	const q = `
		SELECT id, user_id, device_id, token, platform, created_at
		FROM push_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list push tokens: %w", err)
	}
	defer rows.Close()

	var out []Token
	for rows.Next() {
		var t Token
		if err := rows.Scan(&t.ID, &t.UserID, &t.DeviceID, &t.Token, &t.Platform, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan push token: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Handler exposes push token REST endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a push token handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// HandleRegister registers a push token.
// POST /api/push/tokens
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req struct {
		DeviceID string `json:"device_id"`
		Token    string `json:"token"`
		Platform string `json:"platform"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}
	if req.DeviceID == "" || req.Token == "" || req.Platform == "" {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "device_id, token, platform required")
		return
	}

	tok, err := h.repo.Register(r.Context(), claims.UserID, req.DeviceID, req.Token, req.Platform)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to register token")
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, map[string]any{
		"id":         tok.ID,
		"user_id":    tok.UserID,
		"device_id":  tok.DeviceID,
		"platform":   tok.Platform,
		"created_at": tok.CreatedAt.UTC().Format(time.RFC3339),
	})
}

// HandleList returns push tokens for the authenticated user.
// GET /api/push/tokens
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	tokens, err := h.repo.List(r.Context(), claims.UserID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list tokens")
		return
	}

	items := make([]map[string]any, 0, len(tokens))
	for _, t := range tokens {
		items = append(items, map[string]any{
			"id":         t.ID,
			"device_id":  t.DeviceID,
			"platform":   t.Platform,
			"created_at": t.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{"tokens": items})
}
