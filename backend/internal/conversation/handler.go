package conversation

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/cache"
	"github.com/echoline/echoline/backend/internal/entitlement"
)

// Handler exposes conversation REST endpoints.
type Handler struct {
	repo        *Repository
	cache       *cache.ConversationListCache
	entitlement EntitlementGate
}

// EntitlementGate checks paid channel access before subscribe.
type EntitlementGate = entitlement.Gate

// NewHandler creates a conversation handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// SetEntitlementGate enables paid channel subscribe checks.
func (h *Handler) SetEntitlementGate(g EntitlementGate) {
	h.entitlement = g
}

// SetListCache enables Redis-backed conversation list caching.
func (h *Handler) SetListCache(c *cache.ConversationListCache) {
	h.cache = c
}

type directRequest struct {
	UserID string `json:"user_id"`
}

type groupRequest struct {
	Title     string   `json:"title"`
	MemberIDs []string `json:"member_ids"`
}

type channelRequest struct {
	Title string `json:"title"`
}

type memberInviteRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// HandleCreateChannel creates a broadcast channel.
func (h *Handler) HandleCreateChannel(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req channelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	conv, err := h.repo.CreateChannel(r.Context(), claims.UserID, req.Title)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, toConversationResponse(conv))
}

// HandleSubscribe subscribes the current user to a channel.
func (h *Handler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, err := ParseConversationID(r.URL.Path, "/api/conversations/", "/subscribe")
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	if h.entitlement != nil {
		if err := h.entitlement.CanSubscribe(r.Context(), claims.UserID, convID); err != nil {
			if errors.Is(err, entitlement.ErrEntitlementRequired) {
				apierror.Write(w, r, http.StatusPaymentRequired, "payment_required", "paid channel entitlement required")
				return
			}
			apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
	}

	if err := h.repo.Subscribe(r.Context(), convID, claims.UserID); err != nil {
		if errors.Is(err, ErrInvalidType) {
			apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "not a channel")
			return
		}
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]string{"status": "subscribed"})
}

// HandleUnsubscribe removes channel subscription for current user.
func (h *Handler) HandleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, err := ParseConversationID(r.URL.Path, "/api/conversations/", "/subscribe")
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	if err := h.repo.Unsubscribe(r.Context(), convID, claims.UserID); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]string{"status": "unsubscribed"})
}

// HandleInviteMember adds a user to a group.
func (h *Handler) HandleInviteMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	convID, err := ParseConversationID(r.URL.Path, "/api/conversations/", "/members")
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}

	actor, err := h.repo.GetMember(r.Context(), convID, claims.UserID)
	if err != nil || !CanManageMembers(actor.Role) {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "insufficient permissions")
		return
	}

	var req memberInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid user_id")
		return
	}

	role := RoleMember
	if req.Role == string(RoleAdmin) {
		role = RoleAdmin
	}

	if err := h.repo.AddMember(r.Context(), convID, userID, role); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, map[string]string{"status": "invited"})
}

// HandleRemoveMember kicks a user or leaves the group.
func (h *Handler) HandleRemoveMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	// Path: /api/conversations/{id}/members/{user_id}
	path := r.URL.Path
	const prefix = "/api/conversations/"
	if len(path) <= len(prefix)+len("/members/") {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid path")
		return
	}
	rest := path[len(prefix):]
	slash := strings.Index(rest, "/members/")
	if slash < 0 {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid path")
		return
	}
	convID, err := uuid.Parse(rest[:slash])
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid conversation_id")
		return
	}
	targetID, err := uuid.Parse(rest[slash+len("/members/"):])
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid user_id")
		return
	}

	if targetID != claims.UserID {
		actor, err := h.repo.GetMember(r.Context(), convID, claims.UserID)
		if err != nil || !CanManageMembers(actor.Role) {
			apierror.Write(w, r, http.StatusForbidden, "forbidden", "insufficient permissions")
			return
		}
	}

	if err := h.repo.RemoveMember(r.Context(), convID, targetID); err != nil {
		if errors.Is(err, ErrForbidden) {
			apierror.Write(w, r, http.StatusForbidden, "forbidden", "cannot remove owner")
			return
		}
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// HandleCreateDirect creates or returns an existing direct conversation.
func (h *Handler) HandleCreateDirect(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req directRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	peerID, err := uuid.Parse(req.UserID)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid user_id")
		return
	}

	conv, created, err := h.repo.CreateDirect(r.Context(), claims.UserID, peerID)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	apierror.WriteJSON(w, status, toConversationResponse(conv))
}

// HandleCreateGroup creates a group conversation.
func (h *Handler) HandleCreateGroup(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	var req groupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	memberIDs := make([]uuid.UUID, 0, len(req.MemberIDs))
	for _, raw := range req.MemberIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid member_ids")
			return
		}
		memberIDs = append(memberIDs, id)
	}

	conv, err := h.repo.CreateGroup(r.Context(), claims.UserID, req.Title, memberIDs)
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusCreated, toConversationResponse(conv))
}

// HandleList returns conversations for the authenticated user.
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	userKey := claims.UserID.String()
	if h.cache != nil {
		if cached, hit, err := h.cache.Get(r.Context(), userKey); err == nil && hit {
			items := make([]map[string]any, 0, len(cached))
			for _, c := range cached {
				items = append(items, map[string]any{
					"id":         c.ID,
					"type":       c.Type,
					"title":      c.Title,
					"unread":     c.Unread,
					"latest_seq": c.LatestSeq,
				})
			}
			writeJSON(w, r, http.StatusOK, map[string]any{"conversations": items, "cached": true})
			return
		}
	}

	conversations, unreads, err := h.repo.ListForUserWithUnread(r.Context(), claims.UserID, 50)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list conversations")
		return
	}

	items := make([]map[string]any, 0, len(conversations))
	cacheItems := make([]cache.ConversationSummary, 0, len(conversations))
	for i, conv := range conversations {
		item := toConversationResponse(&conv)
		item["unread"] = unreads[i]
		if member, err := h.repo.GetMember(r.Context(), conv.ID, claims.UserID); err == nil {
			item["role"] = member.Role
			item["can_publish"] = CanPublish(conv.Type, member.Role)
		}
		items = append(items, item)
		cacheItems = append(cacheItems, cache.ConversationSummary{
			ID:        conv.ID.String(),
			Type:      string(conv.Type),
			Title:     conv.Title,
			Unread:    unreads[i],
			LatestSeq: conv.LatestSeq,
		})
	}

	if h.cache != nil {
		_ = h.cache.Set(r.Context(), userKey, cacheItems)
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"conversations": items,
	})
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, payload any) {
	apierror.WriteJSON(w, status, payload)
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

// ParseConversationID extracts conversation_id from route patterns like /api/conversations/{id}/...
func ParseConversationID(path, prefix, suffix string) (uuid.UUID, error) {
	if len(path) <= len(prefix)+len(suffix) {
		return uuid.Nil, errors.New("invalid path")
	}
	raw := path[len(prefix) : len(path)-len(suffix)]
	return uuid.Parse(raw)
}
