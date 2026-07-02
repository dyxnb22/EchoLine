package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/user"
	"github.com/echoline/echoline/backend/internal/validate"
)

// LoginAuditor records login attempts.
type LoginAuditor interface {
	LogLogin(ctx context.Context, userID *uuid.UUID, username string, success bool, ip string) error
}

// Service handles authentication flows.
type Service struct {
	users     *user.Repository
	jwtSecret []byte
	auditor   LoginAuditor
}

// NewService creates an auth service.
func NewService(users *user.Repository, jwtSecret string) *Service {
	return &Service{
		users:     users,
		jwtSecret: []byte(jwtSecret),
	}
}

// SetLoginAuditor attaches optional login audit logging.
func (s *Service) SetLoginAuditor(a LoginAuditor) {
	s.auditor = a
}

type registerRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims are JWT claims for authenticated users.
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

type contextKey string

const claimsContextKey contextKey = "auth_claims"

// HandleRegister registers a new user account.
func (s *Service) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r)
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	username, err := validate.Username(req.Username)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	displayName, err := validate.DisplayName(req.DisplayName)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := validate.Password(req.Password); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if displayName == "" {
		displayName = username
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to hash password")
		return
	}

	u, err := s.users.Create(r.Context(), username, displayName, hash)
	if err != nil {
		if errors.Is(err, user.ErrDuplicateUsername) {
			writeError(w, r, http.StatusConflict, "username_taken", "username already exists")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":           u.ID,
		"username":     u.Username,
		"display_name": u.DisplayName,
		"created_at":   u.CreatedAt,
	})
}

// HandleLogin authenticates a user and returns JWT tokens.
func (s *Service) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	u, err := s.users.GetByUsername(r.Context(), req.Username)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			s.auditLogin(r, nil, req.Username, false)
			writeError(w, r, http.StatusUnauthorized, "invalid_credentials", "invalid username or password")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to load user")
		return
	}

	ok, err := VerifyPassword(req.Password, u.PasswordHash)
	if err != nil || !ok {
		s.auditLogin(r, &u.ID, req.Username, false)
		writeError(w, r, http.StatusUnauthorized, "invalid_credentials", "invalid username or password")
		return
	}

	tokens, err := s.issueTokens(u)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to issue token")
		return
	}

	s.auditLogin(r, &u.ID, req.Username, true)
	writeJSON(w, http.StatusOK, tokens)
}

func (s *Service) auditLogin(r *http.Request, userID *uuid.UUID, username string, success bool) {
	if s.auditor == nil {
		return
	}
	ip := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ip = xff
	}
	_ = s.auditor.LogLogin(r.Context(), userID, username, success, ip)
}

// HandleRefresh exchanges a refresh token for new access/refresh tokens.
func (s *Service) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r)
		return
	}

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	claims, err := s.ParseToken(req.RefreshToken)
	if err != nil || claims.TokenType != TokenTypeRefresh {
		writeError(w, r, http.StatusUnauthorized, "invalid_token", "invalid or expired refresh token")
		return
	}

	u, err := s.users.GetByID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "invalid_token", "invalid or expired refresh token")
		return
	}

	tokens, err := s.issueTokens(u)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to issue token")
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

// GetUserByID loads a user for authenticated routes.
func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.users.GetByID(ctx, id)
}

// ParseToken validates a bearer token and returns claims.
func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.TokenType == "" {
		claims.TokenType = TokenTypeAccess
	}
	return claims, nil
}

func (s *Service) issueTokens(u *user.User) (tokenResponse, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)

	accessClaims := Claims{
		UserID:    u.ID,
		Username:  u.Username,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.jwtSecret)
	if err != nil {
		return tokenResponse{}, err
	}

	refreshClaims := accessClaims
	refreshClaims.TokenType = TokenTypeRefresh
	refreshClaims.ExpiresAt = jwt.NewNumericDate(now.Add(7 * 24 * time.Hour))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.jwtSecret)
	if err != nil {
		return tokenResponse{}, err
	}

	return tokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * time.Hour / time.Second),
	}, nil
}

// RequireAuth protects routes with JWT bearer authentication.
func RequireAuth(svc *Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		claims, err := svc.ParseToken(token)
		if err != nil || claims.TokenType != TokenTypeAccess {
			apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ClaimsFromContext extracts JWT claims from request context.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}

// ContextWithClaims attaches claims for tests.
func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	apierror.Write(w, r, status, code, message)
}

func writeMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	apierror.Write(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}
