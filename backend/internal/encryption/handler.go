package encryption

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

// KeyBundle is a device public key registration.
type KeyBundle struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	DeviceID  string
	PublicKey string
	CreatedAt time.Time
}

// Repository persists encryption key bundles.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates an encryption repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Register upserts a device key bundle.
func (r *Repository) Register(ctx context.Context, userID uuid.UUID, deviceID, publicKey string) (*KeyBundle, error) {
	const q = `
		INSERT INTO encryption_key_bundles (id, user_id, device_id, public_key, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
		ON CONFLICT (user_id, device_id) DO UPDATE
			SET public_key = EXCLUDED.public_key, created_at = EXCLUDED.created_at
		RETURNING id, user_id, device_id, public_key, created_at
	`
	now := time.Now().UTC()
	row := r.pool.QueryRow(ctx, q, userID, deviceID, publicKey, now)
	var kb KeyBundle
	if err := row.Scan(&kb.ID, &kb.UserID, &kb.DeviceID, &kb.PublicKey, &kb.CreatedAt); err != nil {
		return nil, fmt.Errorf("register key bundle: %w", err)
	}
	return &kb, nil
}

// List returns key bundles for a user.
func (r *Repository) List(ctx context.Context, userID uuid.UUID) ([]KeyBundle, error) {
	const q = `
		SELECT id, user_id, device_id, public_key, created_at
		FROM encryption_key_bundles
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list key bundles: %w", err)
	}
	defer rows.Close()

	var out []KeyBundle
	for rows.Next() {
		var kb KeyBundle
		if err := rows.Scan(&kb.ID, &kb.UserID, &kb.DeviceID, &kb.PublicKey, &kb.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan key bundle: %w", err)
		}
		out = append(out, kb)
	}
	return out, rows.Err()
}

// Handler exposes E2EE key bundle REST endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates an encryption handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// HandleRegister registers a device public key.
// POST /api/encryption/keys
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req struct {
		DeviceID  string `json:"device_id"`
		PublicKey string `json:"public_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}
	if req.DeviceID == "" || req.PublicKey == "" {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "device_id and public_key required")
		return
	}

	kb, err := h.repo.Register(r.Context(), claims.UserID, req.DeviceID, req.PublicKey)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to register key")
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, bundlePayload(kb))
}

// HandleList lists key bundles for the authenticated user.
// GET /api/encryption/keys
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	bundles, err := h.repo.List(r.Context(), claims.UserID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list keys")
		return
	}

	items := make([]map[string]any, 0, len(bundles))
	for i := range bundles {
		items = append(items, bundlePayload(&bundles[i]))
	}
	apierror.WriteJSON(w, http.StatusOK, map[string]any{"keys": items})
}

func bundlePayload(kb *KeyBundle) map[string]any {
	return map[string]any{
		"id":         kb.ID,
		"device_id":  kb.DeviceID,
		"public_key": kb.PublicKey,
		"created_at": kb.CreatedAt.UTC().Format(time.RFC3339),
	}
}
