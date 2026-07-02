package export

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/conversation"
)

// MessageRow is a minimal message view for export.
type MessageRow struct {
	ID        uuid.UUID
	SenderID  uuid.UUID
	Seq       int64
	Type      string
	Body      string
	Status    string
	CreatedAt time.Time
}

// Repository fetches messages for export.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates an export repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ListAll returns all messages in a conversation (no pagination limit for export).
func (r *Repository) ListAll(ctx context.Context, convID uuid.UUID) ([]MessageRow, error) {
	const q = `
		SELECT id, sender_id, seq, type, body, status, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY seq ASC
	`
	rows, err := r.pool.Query(ctx, q, convID)
	if err != nil {
		return nil, fmt.Errorf("export list: %w", err)
	}
	defer rows.Close()

	var out []MessageRow
	for rows.Next() {
		var m MessageRow
		if err := rows.Scan(&m.ID, &m.SenderID, &m.Seq, &m.Type, &m.Body, &m.Status, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("export scan: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Handler exposes export REST endpoints.
type Handler struct {
	repo          *Repository
	conversations *conversation.Repository
}

// NewHandler creates an export handler.
func NewHandler(repo *Repository, conversations *conversation.Repository) *Handler {
	return &Handler{repo: repo, conversations: conversations}
}

// HandleExport returns all messages of a conversation as JSON.
// GET /api/conversations/{id}/export
func (h *Handler) HandleExport(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	// Parse /api/conversations/{id}/export
	path := r.URL.Path
	const prefix = "/api/conversations/"
	if !strings.HasPrefix(path, prefix) {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid path")
		return
	}
	rest := strings.TrimPrefix(path, prefix)
	idStr := strings.TrimSuffix(rest, "/export")
	convID, err := uuid.Parse(idStr)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	member, err := h.conversations.IsMember(r.Context(), convID, claims.UserID)
	if err != nil || !member {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a conversation member")
		return
	}

	messages, err := h.repo.ListAll(r.Context(), convID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to export messages")
		return
	}

	items := make([]map[string]any, 0, len(messages))
	for _, m := range messages {
		items = append(items, map[string]any{
			"id":         m.ID,
			"sender_id":  m.SenderID,
			"seq":        m.Seq,
			"type":       m.Type,
			"body":       m.Body,
			"status":     m.Status,
			"created_at": m.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="conversation-%s.json"`, convID))
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"conversation_id": convID,
		"exported_at":     time.Now().UTC().Format(time.RFC3339),
		"messages":        items,
	})
}
