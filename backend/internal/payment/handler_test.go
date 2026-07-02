package payment

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/auth"
)

func TestHandleCreateSelfServeDisabled(t *testing.T) {
	h := NewHandler(nil)
	h.SetSelfServe(false)

	body, _ := json.Marshal(map[string]any{
		"amount_cents": 100,
		"currency":     "USD",
		"reference":    "channel:" + uuid.New().String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/payments/ledger", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.Claims{UserID: uuid.New()}))
	rec := httptest.NewRecorder()

	h.HandleCreate(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestHandleSettleSelfServeDisabled(t *testing.T) {
	h := NewHandler(nil)
	h.SetSelfServe(false)

	body, _ := json.Marshal(map[string]string{"reference": "channel:" + uuid.New().String()})
	req := httptest.NewRequest(http.MethodPost, "/api/payments/ledger/settle", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.Claims{UserID: uuid.New()}))
	rec := httptest.NewRecorder()

	h.HandleSettle(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}
