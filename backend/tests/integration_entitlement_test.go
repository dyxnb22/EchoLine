package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIntegrationPaidChannelEntitlement(t *testing.T) {
	handler, _ := setupIntegrationServer(t)

	ownerName := fmt.Sprintf("owner_%d", time.Now().UnixNano())
	ownerToken := registerAndLogin(t, handler, ownerName)

	chBody, _ := json.Marshal(map[string]string{"title": "Paid Channel"})
	chReq := httptest.NewRequest(http.MethodPost, "/api/conversations/channels", bytes.NewReader(chBody))
	chReq.Header.Set("Content-Type", "application/json")
	chReq.Header.Set("Authorization", "Bearer "+ownerToken)
	chW := httptest.NewRecorder()
	handler.ServeHTTP(chW, chReq)
	if chW.Code != http.StatusCreated {
		t.Fatalf("create channel status %d: %s", chW.Code, chW.Body.String())
	}
	var chResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(chW.Body.Bytes(), &chResp); err != nil {
		t.Fatal(err)
	}

	requireBody, _ := json.Marshal(map[string]bool{"required": true})
	requireReq := httptest.NewRequest(http.MethodPost, "/api/channels/"+chResp.ID+"/entitlements/require", bytes.NewReader(requireBody))
	requireReq.Header.Set("Content-Type", "application/json")
	requireReq.Header.Set("Authorization", "Bearer "+ownerToken)
	requireReq.SetPathValue("channel_id", chResp.ID)
	requireW := httptest.NewRecorder()
	handler.ServeHTTP(requireW, requireReq)
	if requireW.Code != http.StatusOK {
		t.Fatalf("set paid status %d: %s", requireW.Code, requireW.Body.String())
	}

	subscriberName := fmt.Sprintf("sub_%d", time.Now().UnixNano())
	subToken := registerAndLogin(t, handler, subscriberName)

	subReq := httptest.NewRequest(http.MethodPost, "/api/conversations/"+chResp.ID+"/subscribe", nil)
	subReq.Header.Set("Authorization", "Bearer "+subToken)
	subReq.SetPathValue("id", chResp.ID)
	subW := httptest.NewRecorder()
	handler.ServeHTTP(subW, subReq)
	if subW.Code != http.StatusPaymentRequired {
		t.Fatalf("subscribe without entitlement status %d, want 402", subW.Code)
	}

	grantBody, _ := json.Marshal(map[string]string{"user_id": "", "reference": "test-grant"})
	grantReq := httptest.NewRequest(http.MethodPost, "/api/channels/"+chResp.ID+"/entitlements/grant", bytes.NewReader(grantBody))
	grantReq.Header.Set("Content-Type", "application/json")
	grantReq.Header.Set("Authorization", "Bearer "+subToken)
	grantReq.SetPathValue("channel_id", chResp.ID)
	grantW := httptest.NewRecorder()
	handler.ServeHTTP(grantW, grantReq)
	if grantW.Code != http.StatusForbidden {
		t.Fatalf("non-admin grant status %d, want 403", grantW.Code)
	}

	settleBody, _ := json.Marshal(map[string]string{"reference": "channel:" + chResp.ID})
	ledgerBody, _ := json.Marshal(map[string]any{
		"amount_cents": 500,
		"currency":     "USD",
		"reference":    "channel:" + chResp.ID,
	})
	ledgerReq := httptest.NewRequest(http.MethodPost, "/api/payments/ledger", bytes.NewReader(ledgerBody))
	ledgerReq.Header.Set("Content-Type", "application/json")
	ledgerReq.Header.Set("Authorization", "Bearer "+subToken)
	ledgerW := httptest.NewRecorder()
	handler.ServeHTTP(ledgerW, ledgerReq)
	if ledgerW.Code != http.StatusCreated {
		t.Fatalf("create ledger status %d: %s", ledgerW.Code, ledgerW.Body.String())
	}

	settleReq := httptest.NewRequest(http.MethodPost, "/api/payments/ledger/settle", bytes.NewReader(settleBody))
	settleReq.Header.Set("Content-Type", "application/json")
	settleReq.Header.Set("Authorization", "Bearer "+subToken)
	settleW := httptest.NewRecorder()
	handler.ServeHTTP(settleW, settleReq)
	if settleW.Code != http.StatusOK && settleW.Code != http.StatusCreated {
		t.Fatalf("settle status %d: %s", settleW.Code, settleW.Body.String())
	}

	subReq2 := httptest.NewRequest(http.MethodPost, "/api/conversations/"+chResp.ID+"/subscribe", nil)
	subReq2.Header.Set("Authorization", "Bearer "+subToken)
	subReq2.SetPathValue("id", chResp.ID)
	subW2 := httptest.NewRecorder()
	handler.ServeHTTP(subW2, subReq2)
	if subW2.Code != http.StatusOK && subW2.Code != http.StatusCreated {
		t.Fatalf("subscribe after payment status %d: %s", subW2.Code, subW2.Body.String())
	}
}
