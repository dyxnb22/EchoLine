package forward

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/message"
)

// Repository handles message forwarding.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a forward repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ErrSourceNotFound is returned when the source message does not exist.
var ErrSourceNotFound = errors.New("source message not found")

// Forward copies a message body to a target conversation.
func (r *Repository) Forward(ctx context.Context, sourceMsgID, targetConvID, senderID uuid.UUID) (*message.Message, error) {
	// Fetch source body.
	const fetchQ = `SELECT body, type FROM messages WHERE id = $1`
	var body string
	var typ message.Type
	row := r.pool.QueryRow(ctx, fetchQ, sourceMsgID)
	if err := row.Scan(&body, &typ); err != nil {
		return nil, fmt.Errorf("forward fetch source: %w", ErrSourceNotFound)
	}

	const q = `
		INSERT INTO messages (id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at)
		VALUES (
			gen_random_uuid(),
			$1, $2,
			'',
			(SELECT COALESCE(MAX(seq), 0) + 1 FROM messages WHERE conversation_id = $1),
			$3, $4,
			'sent',
			$5, $5
		)
		RETURNING id, conversation_id, sender_id, client_msg_id, seq, type, body, status, created_at, updated_at
	`
	now := time.Now().UTC()
	insRow := r.pool.QueryRow(ctx, q, targetConvID, senderID, typ, body, now)
	var msg message.Message
	if err := insRow.Scan(
		&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.ClientMsgID,
		&msg.Seq, &msg.Type, &msg.Body, &msg.Status,
		&msg.CreatedAt, &msg.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("forward insert: %w", err)
	}
	return &msg, nil
}

// Handler exposes forward REST endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a forward handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// HandleForward forwards a message to another conversation.
// POST /api/messages/{message_id}/forward
func (h *Handler) HandleForward(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	// Parse message_id from path /api/messages/{message_id}/forward
	msgIDStr := r.PathValue("message_id")
	if msgIDStr == "" {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "message_id required")
		return
	}
	sourceMsgID, err := uuid.Parse(msgIDStr)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid message_id")
		return
	}

	var req struct {
		TargetConversationID string `json:"target_conversation_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	targetConvID, err := uuid.Parse(req.TargetConversationID)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid target_conversation_id")
		return
	}

	msg, err := h.repo.Forward(r.Context(), sourceMsgID, targetConvID, claims.UserID)
	if err != nil {
		if errors.Is(err, ErrSourceNotFound) {
			apierror.Write(w, r, http.StatusNotFound, "not_found", "source message not found")
			return
		}
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to forward message")
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, map[string]any{
		"id":              msg.ID,
		"conversation_id": msg.ConversationID,
		"sender_id":       msg.SenderID,
		"seq":             msg.Seq,
		"type":            msg.Type,
		"body":            msg.Body,
		"status":          msg.Status,
		"created_at":      msg.CreatedAt.UTC().Format(time.RFC3339),
	})
}
