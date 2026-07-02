package conversation

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

// ArchiveRepository manages conversation archive state per member.
type ArchiveRepository struct {
	pool *pgxpool.Pool
}

// NewArchiveRepository creates an archive repository.
func NewArchiveRepository(pool *pgxpool.Pool) *ArchiveRepository {
	return &ArchiveRepository{pool: pool}
}

// Archive sets archived_at for a member's conversation.
func (r *ArchiveRepository) Archive(ctx context.Context, convID, userID uuid.UUID) error {
	const q = `
		UPDATE conversation_members
		SET archived_at = $3
		WHERE conversation_id = $1 AND user_id = $2
	`
	tag, err := r.pool.Exec(ctx, q, convID, userID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("archive conversation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotMember
	}
	return nil
}

// Unarchive clears archived_at for a member.
func (r *ArchiveRepository) Unarchive(ctx context.Context, convID, userID uuid.UUID) error {
	const q = `
		UPDATE conversation_members
		SET archived_at = NULL
		WHERE conversation_id = $1 AND user_id = $2
	`
	tag, err := r.pool.Exec(ctx, q, convID, userID)
	if err != nil {
		return fmt.Errorf("unarchive conversation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotMember
	}
	return nil
}

// ArchivedConversation is a minimal view of an archived conversation.
type ArchivedConversation struct {
	ConversationID uuid.UUID
	Type           string
	Title          string
	ArchivedAt     time.Time
}

// ListArchived returns all conversations archived by a user.
func (r *ArchiveRepository) ListArchived(ctx context.Context, userID uuid.UUID) ([]ArchivedConversation, error) {
	const q = `
		SELECT cm.conversation_id, c.type, c.title, cm.archived_at
		FROM conversation_members cm
		JOIN conversations c ON c.id = cm.conversation_id
		WHERE cm.user_id = $1 AND cm.archived_at IS NOT NULL
		ORDER BY cm.archived_at DESC
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list archived: %w", err)
	}
	defer rows.Close()

	var out []ArchivedConversation
	for rows.Next() {
		var ac ArchivedConversation
		if err := rows.Scan(&ac.ConversationID, &ac.Type, &ac.Title, &ac.ArchivedAt); err != nil {
			return nil, fmt.Errorf("scan archived: %w", err)
		}
		out = append(out, ac)
	}
	return out, rows.Err()
}

// ArchiveHandler exposes archive REST endpoints.
type ArchiveHandler struct {
	repo    *ArchiveRepository
	convRepo *Repository
}

// NewArchiveHandler creates an archive handler.
func NewArchiveHandler(repo *ArchiveRepository, convRepo *Repository) *ArchiveHandler {
	return &ArchiveHandler{repo: repo, convRepo: convRepo}
}

// HandleArchive archives a conversation for the authenticated user.
// POST /api/conversations/{id}/archive
func (h *ArchiveHandler) HandleArchive(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, err := ParseConversationID(r.URL.Path, "/api/conversations/", "/archive")
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	if err := h.repo.Archive(r.Context(), convID, claims.UserID); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]string{"status": "archived"})
}

// HandleUnarchive unarchives a conversation.
// POST /api/conversations/{id}/unarchive
func (h *ArchiveHandler) HandleUnarchive(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, err := ParseConversationID(r.URL.Path, "/api/conversations/", "/unarchive")
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	if err := h.repo.Unarchive(r.Context(), convID, claims.UserID); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]string{"status": "unarchived"})
}

// HandleListArchived returns all archived conversations for the user.
// GET /api/conversations/archived
func (h *ArchiveHandler) HandleListArchived(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	archived, err := h.repo.ListArchived(r.Context(), claims.UserID)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list archived conversations")
		return
	}

	items := make([]map[string]any, 0, len(archived))
	for _, ac := range archived {
		items = append(items, map[string]any{
			"conversation_id": ac.ConversationID,
			"type":            ac.Type,
			"title":           ac.Title,
			"archived_at":     ac.ArchivedAt.UTC().Format(time.RFC3339),
		})
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{"archived": items})
}
