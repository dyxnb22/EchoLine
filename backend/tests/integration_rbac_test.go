package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIntegrationExportRBAC(t *testing.T) {
	handler, _ := setupIntegrationServer(t)

	owner := registerAndLogin(t, handler, fmt.Sprintf("owner_%d", time.Now().UnixNano()))
	other := registerAndLogin(t, handler, fmt.Sprintf("other_%d", time.Now().UnixNano()))

	groupBody, _ := json.Marshal(map[string]any{"title": "RBAC Group", "member_ids": []string{}})
	groupReq := httptest.NewRequest(http.MethodPost, "/api/conversations/groups", bytes.NewReader(groupBody))
	groupReq.Header.Set("Content-Type", "application/json")
	groupReq.Header.Set("Authorization", "Bearer "+owner)
	groupW := httptest.NewRecorder()
	handler.ServeHTTP(groupW, groupReq)
	if groupW.Code != http.StatusCreated {
		t.Fatalf("create group: %d %s", groupW.Code, groupW.Body.String())
	}
	var group struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(groupW.Body.Bytes(), &group)

	exportReq := httptest.NewRequest(http.MethodGet, "/api/conversations/"+group.ID+"/export", nil)
	exportReq.Header.Set("Authorization", "Bearer "+other)
	exportW := httptest.NewRecorder()
	handler.ServeHTTP(exportW, exportReq)
	if exportW.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-member export, got %d", exportW.Code)
	}
}

func TestIntegrationThreadReplyUsesMainPipeline(t *testing.T) {
	handler, _ := setupIntegrationServer(t)
	token := registerAndLogin(t, handler, fmt.Sprintf("thread_%d", time.Now().UnixNano()))

	groupBody, _ := json.Marshal(map[string]any{"title": "Thread Group", "member_ids": []string{}})
	groupReq := httptest.NewRequest(http.MethodPost, "/api/conversations/groups", bytes.NewReader(groupBody))
	groupReq.Header.Set("Content-Type", "application/json")
	groupReq.Header.Set("Authorization", "Bearer "+token)
	groupW := httptest.NewRecorder()
	handler.ServeHTTP(groupW, groupReq)
	if groupW.Code != http.StatusCreated {
		t.Fatalf("create group: %d", groupW.Code)
	}
	var group struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(groupW.Body.Bytes(), &group)

	parentBody, _ := json.Marshal(map[string]string{
		"body":          "parent",
		"client_msg_id": uuid.New().String(),
	})
	parentReq := httptest.NewRequest(http.MethodPost, "/api/conversations/"+group.ID+"/messages", bytes.NewReader(parentBody))
	parentReq.Header.Set("Content-Type", "application/json")
	parentReq.Header.Set("Authorization", "Bearer "+token)
	parentW := httptest.NewRecorder()
	handler.ServeHTTP(parentW, parentReq)
	if parentW.Code != http.StatusCreated {
		t.Fatalf("send parent: %d %s", parentW.Code, parentW.Body.String())
	}
	var parent struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(parentW.Body.Bytes(), &parent)

	replyBody, _ := json.Marshal(map[string]string{"body": "reply text"})
	replyReq := httptest.NewRequest(http.MethodPost,
		"/api/conversations/"+group.ID+"/messages/"+parent.ID+"/replies",
		bytes.NewReader(replyBody))
	replyReq.Header.Set("Content-Type", "application/json")
	replyReq.Header.Set("Authorization", "Bearer "+token)
	replyW := httptest.NewRecorder()
	handler.ServeHTTP(replyW, replyReq)
	if replyW.Code != http.StatusCreated {
		t.Fatalf("thread reply: %d %s", replyW.Code, replyW.Body.String())
	}
}
