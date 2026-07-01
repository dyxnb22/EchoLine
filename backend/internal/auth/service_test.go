package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/echoline/echoline/backend/internal/migrate"
	"github.com/echoline/echoline/backend/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func openAuthService(t *testing.T) *Service {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	if err := migrate.Up(ctx, dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	return NewService(user.NewRepository(pool), "test-jwt-secret")
}

func TestRegisterAndLogin(t *testing.T) {
	svc := openAuthService(t)
	username := "auth_" + t.Name()

	registerBody := map[string]string{
		"username":     username,
		"password":     "secret123",
		"display_name": "Auth User",
	}
	body, _ := json.Marshal(registerBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	svc.HandleRegister(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", rec.Code, rec.Body.String())
	}

	loginBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": "secret123",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	rec = httptest.NewRecorder()
	svc.HandleLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var tokens tokenResponse
	if err := json.NewDecoder(rec.Body).Decode(&tokens); err != nil {
		t.Fatalf("decode tokens: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Fatal("expected access token")
	}

	claims, err := svc.ParseToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.Username != username {
		t.Fatalf("username = %q, want %q", claims.Username, username)
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	svc := openAuthService(t)
	username := "badpw_" + t.Name()

	registerBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": "secret123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(registerBody))
	rec := httptest.NewRecorder()
	svc.HandleRegister(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register status = %d", rec.Code)
	}

	loginBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": "wrong",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	rec = httptest.NewRecorder()
	svc.HandleLogin(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthMiddleware(t *testing.T) {
	svc := openAuthService(t)

	protected := RequireAuth(svc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	username := "mw_" + t.Name()
	registerBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": "secret123",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(registerBody))
	rec = httptest.NewRecorder()
	svc.HandleRegister(rec, req)

	loginBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": "secret123",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	rec = httptest.NewRecorder()
	svc.HandleLogin(rec, req)

	var tokens tokenResponse
	_ = json.NewDecoder(rec.Body).Decode(&tokens)

	req = httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec = httptest.NewRecorder()
	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
