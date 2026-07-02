package entitlement

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/admin"
	"github.com/echoline/echoline/backend/internal/auth"
)

type stubOwner struct {
	owner bool
	err   error
}

func (s stubOwner) IsChannelOwner(_ context.Context, _, _ uuid.UUID) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.owner, nil
}

func withClaims(r *http.Request, userID uuid.UUID) *http.Request {
	claims := &auth.Claims{UserID: userID, Username: "tester"}
	return r.WithContext(auth.ContextWithClaims(r.Context(), claims))
}

func TestHandleSetPaidRequiresOwner(t *testing.T) {
	otherID := uuid.New()
	channelID := uuid.New()

	h := NewHandler(nil, stubOwner{owner: false}, nil)

	body, _ := json.Marshal(map[string]bool{"required": true})
	req := httptest.NewRequest(http.MethodPost, "/api/channels/"+channelID.String()+"/entitlements/require", bytes.NewReader(body))
	req.SetPathValue("channel_id", channelID.String())
	req = withClaims(req, otherID)

	w := httptest.NewRecorder()
	h.HandleSetPaid(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("non-owner status = %d", w.Code)
	}
}

func TestHandleGrantRequiresAdmin(t *testing.T) {
	userID := uuid.New()
	channelID := uuid.New()
	checker := admin.NewStaticAdminChecker(uuid.New().String())
	h := NewHandler(nil, nil, checker)

	body, _ := json.Marshal(map[string]string{"reference": "manual"})
	req := httptest.NewRequest(http.MethodPost, "/api/channels/"+channelID.String()+"/entitlements/grant", bytes.NewReader(body))
	req.SetPathValue("channel_id", channelID.String())
	req = withClaims(req, userID)

	w := httptest.NewRecorder()
	h.HandleGrant(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("non-admin grant status = %d, want 403", w.Code)
	}
}
