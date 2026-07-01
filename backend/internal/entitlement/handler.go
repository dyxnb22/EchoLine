package entitlement

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// Handler exposes entitlement REST endpoints.
type Handler struct {
	repo *Repository
}

// NewHandler creates an entitlement handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// HandleGrant grants channel access after payment.
// POST /api/channels/{channel_id}/entitlements/grant
func (h *Handler) HandleGrant(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Reference string `json:"reference"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.repo.Grant(r.Context(), claims.UserID, channelID, req.Reference); err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to grant entitlement")
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]string{"status": "granted"})
}

// HandleSetPaid marks a channel as requiring entitlement (channel owner).
// POST /api/channels/{channel_id}/entitlements/require
func (h *Handler) HandleSetPaid(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	channelID, err := uuid.Parse(r.PathValue("channel_id"))
	if err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid channel_id")
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
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{"requires_entitlement": req.Required})
}
