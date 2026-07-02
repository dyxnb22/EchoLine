package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"log/slog"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/config"
	"github.com/echoline/echoline/backend/internal/db"
	"github.com/echoline/echoline/backend/internal/migrate"
	"github.com/echoline/echoline/backend/internal/server"
)

func setupIntegrationServer(t *testing.T) (http.Handler, string) {
	t.Helper()
	if !integrationEnabled() {
		t.Skip("integration tests require RUN_INTEGRATION=1 and DATABASE_URL")
	}

	t.Setenv("PAYMENT_SELF_SERVE", "true")

	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if err := migrate.Up(ctx, dbURL); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	pool, err := db.Connect(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	srv := server.New(cfg, pool, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	return srv.Handler(), dbURL
}

func registerAndLogin(t *testing.T, handler http.Handler, username string) string {
	t.Helper()

	regBody, _ := json.Marshal(map[string]string{
		"username":     username,
		"password":     "secret12345",
		"display_name": "Integration User",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	handler.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusCreated {
		t.Fatalf("register status %d: %s", regW.Code, regW.Body.String())
	}

	loginBody, _ := json.Marshal(map[string]string{"username": username, "password": "secret12345"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	handler.ServeHTTP(loginW, loginReq)
	if loginW.Code != http.StatusOK {
		t.Fatalf("login status %d: %s", loginW.Code, loginW.Body.String())
	}

	var tokens struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(loginW.Body.Bytes(), &tokens); err != nil {
		t.Fatal(err)
	}
	if tokens.AccessToken == "" {
		t.Fatal("empty access token")
	}
	return tokens.AccessToken
}

func TestIntegrationMessagingFlow(t *testing.T) {
	handler, _ := setupIntegrationServer(t)

	username := fmt.Sprintf("msguser_%d", time.Now().UnixNano())
	token := registerAndLogin(t, handler, username)

	groupBody, _ := json.Marshal(map[string]any{
		"title":      "Integration Group",
		"member_ids": []string{},
	})
	groupReq := httptest.NewRequest(http.MethodPost, "/api/conversations/groups", bytes.NewReader(groupBody))
	groupReq.Header.Set("Content-Type", "application/json")
	groupReq.Header.Set("Authorization", "Bearer "+token)
	groupW := httptest.NewRecorder()
	handler.ServeHTTP(groupW, groupReq)
	if groupW.Code != http.StatusCreated {
		t.Fatalf("create group status %d: %s", groupW.Code, groupW.Body.String())
	}

	var groupResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(groupW.Body.Bytes(), &groupResp); err != nil {
		t.Fatal(err)
	}
	if groupResp.ID == "" {
		t.Fatal("empty group id")
	}

	msgBody, _ := json.Marshal(map[string]string{
		"body":          "hello integration test",
		"client_msg_id": uuid.New().String(),
	})
	sendReq := httptest.NewRequest(http.MethodPost, "/api/conversations/"+groupResp.ID+"/messages", bytes.NewReader(msgBody))
	sendReq.Header.Set("Content-Type", "application/json")
	sendReq.Header.Set("Authorization", "Bearer "+token)
	sendW := httptest.NewRecorder()
	handler.ServeHTTP(sendW, sendReq)
	if sendW.Code != http.StatusCreated {
		t.Fatalf("send message status %d: %s", sendW.Code, sendW.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/conversations/"+groupResp.ID+"/messages?limit=50", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listW := httptest.NewRecorder()
	handler.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("list messages status %d: %s", listW.Code, listW.Body.String())
	}

	var page struct {
		Messages []struct {
			Body string `json:"body"`
			Seq  int64  `json:"seq"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(listW.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Messages) == 0 {
		t.Fatal("expected at least one message")
	}
	found := false
	for _, m := range page.Messages {
		if m.Body == "hello integration test" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("message body not found in %+v", page.Messages)
	}

	convReq := httptest.NewRequest(http.MethodGet, "/api/conversations", nil)
	convReq.Header.Set("Authorization", "Bearer "+token)
	convW := httptest.NewRecorder()
	handler.ServeHTTP(convW, convReq)
	if convW.Code != http.StatusOK {
		t.Fatalf("list conversations status %d", convW.Code)
	}
}
