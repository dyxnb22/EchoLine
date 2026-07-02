package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/echoline/echoline/backend/internal/validate"
)

func TestIntegrationMessageBodyValidation(t *testing.T) {
	handler, _ := setupIntegrationServer(t)

	username := fmt.Sprintf("valid_%d", time.Now().UnixNano())
	token := registerAndLogin(t, handler, username)

	groupBody, _ := json.Marshal(map[string]any{
		"title":      "Validation Group",
		"member_ids": []string{},
	})
	groupReq := httptest.NewRequest(http.MethodPost, "/api/conversations/groups", bytes.NewReader(groupBody))
	groupReq.Header.Set("Content-Type", "application/json")
	groupReq.Header.Set("Authorization", "Bearer "+token)
	groupW := httptest.NewRecorder()
	handler.ServeHTTP(groupW, groupReq)
	if groupW.Code != http.StatusCreated {
		t.Fatalf("create group status %d", groupW.Code)
	}
	var groupResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(groupW.Body.Bytes(), &groupResp); err != nil {
		t.Fatal(err)
	}

	emptyBody, _ := json.Marshal(map[string]string{
		"body":          "",
		"client_msg_id": "empty-body",
	})
	emptyReq := httptest.NewRequest(http.MethodPost, "/api/conversations/"+groupResp.ID+"/messages", bytes.NewReader(emptyBody))
	emptyReq.Header.Set("Content-Type", "application/json")
	emptyReq.Header.Set("Authorization", "Bearer "+token)
	emptyW := httptest.NewRecorder()
	handler.ServeHTTP(emptyW, emptyReq)
	if emptyW.Code != http.StatusBadRequest {
		t.Fatalf("empty body status %d, want 400", emptyW.Code)
	}

	longBody, _ := json.Marshal(map[string]string{
		"body":          strings.Repeat("x", validate.MaxMessageBodyLen+1),
		"client_msg_id": "long-body",
	})
	longReq := httptest.NewRequest(http.MethodPost, "/api/conversations/"+groupResp.ID+"/messages", bytes.NewReader(longBody))
	longReq.Header.Set("Content-Type", "application/json")
	longReq.Header.Set("Authorization", "Bearer "+token)
	longW := httptest.NewRecorder()
	handler.ServeHTTP(longW, longReq)
	if longW.Code != http.StatusBadRequest {
		t.Fatalf("long body status %d, want 400", longW.Code)
	}
}
