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

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/config"
	"github.com/echoline/echoline/backend/internal/db"
	"github.com/echoline/echoline/backend/internal/migrate"
	"github.com/echoline/echoline/backend/internal/server"
)

func integrationEnabled() bool {
	return os.Getenv("RUN_INTEGRATION") == "1" && os.Getenv("DATABASE_URL") != ""
}

func TestIntegrationAuthRegisterLogin(t *testing.T) {
	if !integrationEnabled() {
		t.Skip("integration tests require RUN_INTEGRATION=1 and DATABASE_URL")
	}
	ensureIntegrationEnv(t)

	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if err := migrate.Up(ctx, dbURL); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	pool, err := db.Connect(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	srv := server.New(cfg, pool, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	handler := srv.Handler()

	username := fmt.Sprintf("intuser_%d", time.Now().UnixNano())
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

	meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	meW := httptest.NewRecorder()
	handler.ServeHTTP(meW, meReq)
	if meW.Code != http.StatusOK {
		t.Fatalf("me status %d", meW.Code)
	}
}

// Ensure pool type used for future tests.
var _ *pgxpool.Pool
var _ *auth.Service
