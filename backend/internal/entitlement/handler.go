package entitlement

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/admin"
	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// Handler exposes entitlement REST endpoints.
type Handler struct {
	repo  *Repository
	owner OwnerChecker
	admin admin.AdminChecker
}

// NewHandler creates an entitlement handler.
func NewHandler(repo *Repository, owner OwnerChecker, adminChecker admin.AdminChecker) *Handler {
	return &Handler{repo: repo, owner: owner, admin: adminChecker}
}

// HandleGrant grants channel access (admin only).
// POST /api/channels/{channel_id}/entitlements/grant
func (h *Handler) HandleGrant(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}
	if h.admin == nil || !h.admin.IsAdmin(claims.UserID) {
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "admin access required")
		return
	}

	channelID, err := uuid.Parse(r.PathValue("channel_id"))
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid channel_id")
		return
	}

	var req struct {
		UserID    string `json:"user_id"`
		Reference string `json:"reference"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}

	targetUser := claims.UserID
	if req.UserID != "" {
		parsed, err := uuid.Parse(req.UserID)
		if err != nil {
			apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid user_id")
			return
		}
		targetUser = parsed
	}

	if err := h.repo.Grant(r.Context(), targetUser, channelID, req.Reference); err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to grant entitlement")
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]string{"status": "granted"})
}

// HandleSetPaid marks a channel as requiring entitlement (channel owner only).
// POST /api/channels/{channel_id}/entitlements/require
func (h *Handler) HandleSetPaid(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	channelID, err := uuid.Parse(r.PathValue("channel_id"))
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid channel_id")
		return
	}

	if err := h.requireChannelOwner(r.Context(), channelID, claims.UserID); err != nil {
		if errors.Is(err, errNotChannelOwner) {
			apierror.Write(w, r, http.StatusForbidden, "forbidden", "channel owner required")
			return
		}
		apierror.Write(w, r, http.StatusForbidden, "forbidden", "not a channel member")
		return
	}

	var req struct {
		Required bool `json:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}

	if err := h.repo.SetChannelRequiresEntitlement(r.Context(), channelID, req.Required); err != nil {
		if errors.Is(err, ErrNotPaidChannel) {
			apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "not a channel conversation")
			return
		}
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{"requires_entitlement": req.Required})
}

var errNotChannelOwner = errors.New("channel owner required")

func (h *Handler) requireChannelOwner(ctx context.Context, channelID, userID uuid.UUID) error {
	if h.owner == nil {
		return errNotChannelOwner
	}
	ok, err := h.owner.IsChannelOwner(ctx, channelID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errNotChannelOwner
	}
	return nil
}
