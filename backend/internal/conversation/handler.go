package conversation

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/auth"
)

// Handler exposes conversation REST endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates a conversation handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

type directRequest struct {
	UserID string `json:"user_id"`
}

type groupRequest struct {
	Title     string   `json:"title"`
	MemberIDs []string `json:"member_ids"`
}

// HandleCreateDirect creates or returns an existing direct conversation.
func (h *Handler) HandleCreateDirect(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req directRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	peerID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid user_id")
		return
	}

	conv, created, err := h.repo.CreateDirect(r.Context(), claims.UserID, peerID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, toConversationResponse(conv))
}

// HandleCreateGroup creates a group conversation.
func (h *Handler) HandleCreateGroup(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req groupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	memberIDs := make([]uuid.UUID, 0, len(req.MemberIDs))
	for _, raw := range req.MemberIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid member_ids")
			return
		}
		memberIDs = append(memberIDs, id)
	}

	conv, err := h.repo.CreateGroup(r.Context(), claims.UserID, req.Title, memberIDs)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toConversationResponse(conv))
}

// HandleList returns conversations for the authenticated user.
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	conversations, unreads, err := h.repo.ListForUserWithUnread(r.Context(), claims.UserID, 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list conversations")
		return
	}

	items := make([]map[string]any, 0, len(conversations))
	for i, conv := range conversations {
		item := toConversationResponse(&conv)
		item["unread"] = unreads[i]
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"conversations": items,
	})
}

func toConversationResponse(conv *Conversation) map[string]any {
	return map[string]any{
		"id":              conv.ID,
		"type":            conv.Type,
		"title":           conv.Title,
		"latest_seq":      conv.LatestSeq,
		"last_message_id": conv.LastMessageID,
		"created_by":      conv.CreatedBy,
		"created_at":      conv.CreatedAt,
		"updated_at":      conv.UpdatedAt,
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// ParseConversationID extracts conversation_id from route patterns like /api/conversations/{id}/...
func ParseConversationID(path, prefix, suffix string) (uuid.UUID, error) {
	if len(path) <= len(prefix)+len(suffix) {
		return uuid.Nil, errors.New("invalid path")
	}
	raw := path[len(prefix) : len(path)-len(suffix)]
	return uuid.Parse(raw)
}
