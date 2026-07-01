package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/config"
	"github.com/echoline/echoline/backend/internal/user"
)

// New creates a new API server.
func New(cfg config.Config, pool *pgxpool.Pool, logger *slog.Logger) *Server {
	return NewWithOptions(cfg, pool, logger, nil)
}

// Handler returns the root HTTP handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /api/auth/register", s.auth.HandleRegister)
	mux.HandleFunc("POST /api/auth/refresh", s.auth.HandleRefresh)
	mux.Handle("GET /api/me", auth.RequireAuth(s.auth, http.HandlerFunc(s.handleMe)))
	mux.Handle("GET /api/conversations", auth.RequireAuth(s.auth, http.HandlerFunc(s.conv.HandleList)))
	mux.Handle("POST /api/conversations/direct", auth.RequireAuth(s.auth, http.HandlerFunc(s.conv.HandleCreateDirect)))
	mux.Handle("POST /api/conversations/groups", auth.RequireAuth(s.auth, http.HandlerFunc(s.conv.HandleCreateGroup)))
	mux.Handle("POST /api/conversations/channels", auth.RequireAuth(s.auth, http.HandlerFunc(s.conv.HandleCreateChannel)))
	mux.Handle("POST /api/conversations/{id}/subscribe", auth.RequireAuth(s.auth, http.HandlerFunc(s.conv.HandleSubscribe)))
	mux.Handle("DELETE /api/conversations/{id}/subscribe", auth.RequireAuth(s.auth, http.HandlerFunc(s.conv.HandleUnsubscribe)))
	mux.Handle("POST /api/conversations/{id}/members", auth.RequireAuth(s.auth, http.HandlerFunc(s.conv.HandleInviteMember)))
	mux.Handle("DELETE /api/conversations/{id}/members/{user_id}", auth.RequireAuth(s.auth, http.HandlerFunc(s.conv.HandleRemoveMember)))
	mux.Handle("GET /api/conversations/{id}/messages", auth.RequireAuth(s.auth, http.HandlerFunc(s.msg.HandleList)))
	mux.Handle("POST /api/conversations/{id}/read", auth.RequireAuth(s.auth, http.HandlerFunc(s.msg.HandleMarkRead)))
	mux.Handle("POST /api/sync", auth.RequireAuth(s.auth, http.HandlerFunc(s.sync.HandleSync)))
	mux.Handle("POST /api/messages/ack", auth.RequireAuth(s.auth, http.HandlerFunc(s.delivery.HandleACK)))
	mux.HandleFunc("GET /ws", s.realtime.HandleWS)

	s.applyRateLimits(mux)

	return apierror.RequestIDMiddleware(s.withLogging(mux))
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"request_id", apierror.RequestIDFromContext(r.Context()),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	status := "ok"
	code := http.StatusOK

	if s.pool == nil {
		status = "degraded"
		code = http.StatusServiceUnavailable
	} else if err := s.pool.Ping(ctx); err != nil {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}

	apierror.WriteJSON(w, code, map[string]string{
		"status": status,
		"env":    s.cfg.AppEnv,
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth context")
		return
	}

	u, err := s.auth.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			apierror.Write(w, r, http.StatusNotFound, "not_found", "user not found")
			return
		}
		s.logger.Error("get user", "error", err)
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to load user")
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"id":           u.ID,
		"username":     u.Username,
		"display_name": u.DisplayName,
		"created_at":   u.CreatedAt,
	})
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	s.httpServer = &http.Server{
		Addr:              s.cfg.HTTPAddr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	s.logger.Info("starting api server", "addr", s.cfg.HTTPAddr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

// Keep writeJSON for tests that still reference server package helpers.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
