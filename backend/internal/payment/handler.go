package payment

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

// LedgerEntry is a payment ledger record.
type LedgerEntry struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	AmountCents int64
	Currency    string
	Status      string
	Reference   string
	CreatedAt   time.Time
}

// Repository persists payment ledger entries.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a payment repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new ledger entry.
func (r *Repository) Create(ctx context.Context, userID uuid.UUID, amountCents int64, currency, status, reference string) (*LedgerEntry, error) {
	const q = `
		INSERT INTO payment_ledger (id, user_id, amount_cents, currency, status, reference, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, amount_cents, currency, status, reference, created_at
	`
	now := time.Now().UTC()
	row := r.pool.QueryRow(ctx, q, userID, amountCents, currency, status, reference, now)
	var e LedgerEntry
	if err := row.Scan(&e.ID, &e.UserID, &e.AmountCents, &e.Currency, &e.Status, &e.Reference, &e.CreatedAt); err != nil {
		return nil, fmt.Errorf("create ledger entry: %w", err)
	}
	return &e, nil
}

// List returns ledger entries for a user.
func (r *Repository) List(ctx context.Context, userID uuid.UUID) ([]LedgerEntry, error) {
	const q = `
		SELECT id, user_id, amount_cents, currency, status, reference, created_at
		FROM payment_ledger
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list ledger: %w", err)
	}
	defer rows.Close()

	var out []LedgerEntry
	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.AmountCents, &e.Currency, &e.Status, &e.Reference, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan ledger: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Settle marks a ledger entry as settled (idempotent by reference).
func (r *Repository) Settle(ctx context.Context, userID uuid.UUID, reference string) (*LedgerEntry, error) {
	const q = `
		UPDATE payment_ledger
		SET status = 'settled'
		WHERE user_id = $1 AND reference = $2 AND status = 'pending'
		RETURNING id, user_id, amount_cents, currency, status, reference, created_at
	`
	row := r.pool.QueryRow(ctx, q, userID, reference)
	var e LedgerEntry
	if err := row.Scan(&e.ID, &e.UserID, &e.AmountCents, &e.Currency, &e.Status, &e.Reference, &e.CreatedAt); err != nil {
		return nil, fmt.Errorf("settle ledger: %w", err)
	}
	return &e, nil
}

// Handler exposes payment ledger REST endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a payment handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// HandleCreate creates a ledger entry (skeleton).
// POST /api/payments/ledger
func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req struct {
		AmountCents int64  `json:"amount_cents"`
		Currency    string `json:"currency"`
		Reference   string `json:"reference"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}

	entry, err := h.repo.Create(r.Context(), claims.UserID, req.AmountCents, req.Currency, "pending", req.Reference)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to create ledger entry")
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, ledgerPayload(entry))
}

// HandleList returns ledger entries for the authenticated user.
// GET /api/payments/ledger
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	entries, err := h.repo.List(r.Context(), claims.UserID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list ledger")
		return
	}

	items := make([]map[string]any, 0, len(entries))
	for i := range entries {
		items = append(items, ledgerPayload(&entries[i]))
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{"entries": items})
}

// HandleSettle settles a pending ledger entry by reference (idempotent).
// POST /api/payments/ledger/settle
func (h *Handler) HandleSettle(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req struct {
		Reference string `json:"reference"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Reference == "" {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "reference is required")
		return
	}

	entry, err := h.repo.Settle(r.Context(), claims.UserID, req.Reference)
	if err != nil {
		apierror.Write(w, r, http.StatusNotFound, "not_found", "pending entry not found")
		return
	}

	apierror.WriteJSON(w, http.StatusOK, ledgerPayload(entry))
}

func ledgerPayload(e *LedgerEntry) map[string]any {
	return map[string]any{
		"id":           e.ID,
		"user_id":      e.UserID,
		"amount_cents": e.AmountCents,
		"currency":     e.Currency,
		"status":       e.Status,
		"reference":    e.Reference,
		"created_at":   e.CreatedAt.UTC().Format(time.RFC3339),
	}
}
