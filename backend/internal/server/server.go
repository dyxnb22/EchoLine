package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/admin"
	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/config"
	"github.com/echoline/echoline/backend/internal/metrics"
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
	mux.Handle("GET /metrics", metrics.ProtectedHandler(s.cfg.MetricsToken))
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
	mux.Handle("PATCH /api/conversations/{id}/messages/{message_id}", auth.RequireAuth(s.auth, http.HandlerFunc(s.msg.HandleEdit)))
	mux.Handle("POST /api/conversations/{id}/messages/{message_id}/recall", auth.RequireAuth(s.auth, http.HandlerFunc(s.msg.HandleRecall)))
	mux.Handle("POST /api/conversations/{id}/read", auth.RequireAuth(s.auth, http.HandlerFunc(s.msg.HandleMarkRead)))
	mux.Handle("POST /api/sync", auth.RequireAuth(s.auth, http.HandlerFunc(s.sync.HandleSync)))
	mux.Handle("GET /api/search/messages", auth.RequireAuth(s.auth, http.HandlerFunc(s.search.HandleSearch)))
	mux.Handle("POST /api/messages/ack", auth.RequireAuth(s.auth, http.HandlerFunc(s.delivery.HandleACK)))
	if s.media != nil {
		mux.Handle("POST /api/media/upload-url", auth.RequireAuth(s.auth, http.HandlerFunc(s.media.HandlePresignUpload)))
		mux.Handle("POST /api/media/download-url", auth.RequireAuth(s.auth, http.HandlerFunc(s.media.HandlePresignDownload)))
	}
	mux.HandleFunc("GET /ws", s.realtime.HandleWS)

	// Profile
	mux.Handle("PATCH /api/me", auth.RequireAuth(s.auth, http.HandlerFunc(s.handlePatchMe)))

	// Devices
	mux.Handle("GET /api/devices", auth.RequireAuth(s.auth, http.HandlerFunc(s.deviceH.HandleList)))

	// Pin / Unpin / List
	mux.Handle("POST /api/conversations/{id}/pins/{message_id}", auth.RequireAuth(s.auth, http.HandlerFunc(s.pin.HandlePin)))
	mux.Handle("DELETE /api/conversations/{id}/pins/{message_id}", auth.RequireAuth(s.auth, http.HandlerFunc(s.pin.HandleUnpin)))
	mux.Handle("GET /api/conversations/{id}/pins", auth.RequireAuth(s.auth, http.HandlerFunc(s.pin.HandleList)))

	// Mute / Unmute
	mux.Handle("POST /api/conversations/{id}/mute", auth.RequireAuth(s.auth, http.HandlerFunc(s.mute.HandleMute)))
	mux.Handle("POST /api/conversations/{id}/unmute", auth.RequireAuth(s.auth, http.HandlerFunc(s.mute.HandleUnmute)))

	// Block / Unblock / List
	mux.Handle("POST /api/blocks/{user_id}", auth.RequireAuth(s.auth, http.HandlerFunc(s.block.HandleBlock)))
	mux.Handle("DELETE /api/blocks/{user_id}", auth.RequireAuth(s.auth, http.HandlerFunc(s.block.HandleUnblock)))
	mux.Handle("GET /api/blocks", auth.RequireAuth(s.auth, http.HandlerFunc(s.block.HandleList)))

	// Reports
	mux.Handle("POST /api/conversations/{id}/messages/{message_id}/report", auth.RequireAuth(s.auth, http.HandlerFunc(s.report.HandleCreate)))

	// Notifications
	mux.Handle("GET /api/notifications", auth.RequireAuth(s.auth, http.HandlerFunc(s.notification.HandleList)))
	mux.Handle("POST /api/notifications/{id}/read", auth.RequireAuth(s.auth, http.HandlerFunc(s.notification.HandleMarkRead)))
	mux.Handle("POST /api/notifications/read-all", auth.RequireAuth(s.auth, http.HandlerFunc(s.notification.HandleMarkAllRead)))

	// Admin
	adminMW := func(h http.Handler) http.Handler {
		return admin.RequireAdmin(s.auth, s.adminChecker, h)
	}
	mux.Handle("GET /api/admin/health", auth.RequireAuth(s.auth, adminMW(http.HandlerFunc(s.adminHandler.HandleHealth))))
	mux.Handle("GET /api/admin/users", auth.RequireAuth(s.auth, adminMW(http.HandlerFunc(s.adminHandler.HandleListUsers))))
	mux.Handle("GET /api/admin/reports", auth.RequireAuth(s.auth, adminMW(http.HandlerFunc(s.adminHandler.HandleListReports))))
	mux.Handle("GET /api/admin/audit-logs", auth.RequireAuth(s.auth, adminMW(http.HandlerFunc(s.adminHandler.HandleListAuditLogs))))
	mux.Handle("GET /api/admin/dlq", auth.RequireAuth(s.auth, adminMW(http.HandlerFunc(s.dlqHandler.HandleList))))
	mux.Handle("POST /api/admin/dlq/{id}/replay", auth.RequireAuth(s.auth, adminMW(http.HandlerFunc(s.dlqReplay.HandleReplay))))

	// Reactions
	mux.Handle("POST /api/messages/{message_id}/reactions", auth.RequireAuth(s.auth, http.HandlerFunc(s.reaction.HandleAdd)))
	mux.Handle("DELETE /api/messages/{message_id}/reactions/{emoji}", auth.RequireAuth(s.auth, http.HandlerFunc(s.reaction.HandleRemove)))
	mux.Handle("GET /api/messages/{message_id}/reactions", auth.RequireAuth(s.auth, http.HandlerFunc(s.reaction.HandleList)))

	// Threads (replies)
	mux.Handle("POST /api/conversations/{conv_id}/messages/{message_id}/replies", auth.RequireAuth(s.auth, http.HandlerFunc(s.thread.HandleSendReply)))
	mux.Handle("GET /api/conversations/{conv_id}/messages/{message_id}/replies", auth.RequireAuth(s.auth, http.HandlerFunc(s.thread.HandleListReplies)))

	// Forward
	mux.Handle("POST /api/messages/{message_id}/forward", auth.RequireAuth(s.auth, http.HandlerFunc(s.forward.HandleForward)))

	// Presence
	mux.Handle("GET /api/presence/online", auth.RequireAuth(s.auth, http.HandlerFunc(s.presenceH.HandleOnline)))
	mux.Handle("GET /api/presence/last-seen", auth.RequireAuth(s.auth, http.HandlerFunc(s.lastSeenH.HandleGet)))
	mux.Handle("POST /api/presence/last-seen", auth.RequireAuth(s.auth, http.HandlerFunc(s.lastSeenH.HandleTouch)))

	// Encryption key bundles
	mux.Handle("POST /api/encryption/keys", auth.RequireAuth(s.auth, http.HandlerFunc(s.encryption.HandleRegister)))
	mux.Handle("GET /api/encryption/keys", auth.RequireAuth(s.auth, http.HandlerFunc(s.encryption.HandleList)))

	// Export
	mux.Handle("GET /api/conversations/{id}/export", auth.RequireAuth(s.auth, http.HandlerFunc(s.export.HandleExport)))

	// Archive
	mux.Handle("POST /api/conversations/{id}/archive", auth.RequireAuth(s.auth, http.HandlerFunc(s.archive.HandleArchive)))
	mux.Handle("POST /api/conversations/{id}/unarchive", auth.RequireAuth(s.auth, http.HandlerFunc(s.archive.HandleUnarchive)))
	mux.Handle("GET /api/conversations/archived", auth.RequireAuth(s.auth, http.HandlerFunc(s.archive.HandleListArchived)))

	// Push tokens
	mux.Handle("POST /api/push/tokens", auth.RequireAuth(s.auth, http.HandlerFunc(s.push.HandleRegister)))
	mux.Handle("GET /api/push/tokens", auth.RequireAuth(s.auth, http.HandlerFunc(s.push.HandleList)))

	// Payments
	mux.Handle("POST /api/payments/ledger", auth.RequireAuth(s.auth, http.HandlerFunc(s.payment.HandleCreate)))
	mux.Handle("GET /api/payments/ledger", auth.RequireAuth(s.auth, http.HandlerFunc(s.payment.HandleList)))
	mux.Handle("POST /api/payments/ledger/settle", auth.RequireAuth(s.auth, http.HandlerFunc(s.payment.HandleSettle)))

	// Ads
	mux.Handle("POST /api/channels/{channel_id}/campaigns", auth.RequireAuth(s.auth, http.HandlerFunc(s.ads.HandleCreate)))
	mux.Handle("GET /api/channels/{channel_id}/campaigns", auth.RequireAuth(s.auth, http.HandlerFunc(s.ads.HandleList)))
	mux.Handle("POST /api/channels/{channel_id}/campaigns/{campaign_id}/impressions", auth.RequireAuth(s.auth, http.HandlerFunc(s.ads.HandleRecordImpression)))

	// GraphQL prototype (POST registered in applyRateLimits with rate limit)
	if s.graph != nil {
		mux.Handle("GET /graphql", auth.RequireAuth(s.auth, http.HandlerFunc(s.graph.HandleGraphQL)))
	}

	// Recommendations
	mux.Handle("GET /api/recommendations/channels", auth.RequireAuth(s.auth, http.HandlerFunc(s.recommendation.HandleRecommendChannels)))
	mux.Handle("GET /api/recommendations/friends", auth.RequireAuth(s.auth, http.HandlerFunc(s.recommendation.HandleRecommendFriends)))

	// Channel entitlements — grant is admin-only; require is owner-only (enforced in handler)
	mux.Handle("POST /api/channels/{channel_id}/entitlements/grant", auth.RequireAuth(s.auth, http.HandlerFunc(s.entitlement.HandleGrant)))
	mux.Handle("POST /api/channels/{channel_id}/entitlements/require", auth.RequireAuth(s.auth, http.HandlerFunc(s.entitlement.HandleSetPaid)))

	s.applyRateLimits(mux)

	return metrics.TraceMiddleware(metrics.HTTPMiddleware(apierror.RequestIDMiddleware(s.withLogging(mux))))
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"request_id", apierror.RequestIDFromContext(r.Context()),
			"trace_id", apierror.TraceIDFromContext(r.Context()),
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

type patchMeRequest struct {
	DisplayName string `json:"display_name"`
}

func (s *Server) handlePatchMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth context")
		return
	}

	var req patchMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.DisplayName == "" {
		apierror.Write(w, r, http.StatusBadRequest, "invalid_request", "display_name is required")
		return
	}

	if err := s.profileRepo.UpdateDisplayName(r.Context(), claims.UserID, req.DisplayName); err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to update profile")
		return
	}

	u, err := s.userRepo.GetByID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			apierror.Write(w, r, http.StatusNotFound, "not_found", "user not found")
			return
		}
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to load profile")
		return
	}

	apierror.WriteJSON(w, http.StatusOK, map[string]any{
		"id":           u.ID,
		"username":     u.Username,
		"display_name": u.DisplayName,
		"updated_at":   u.UpdatedAt,
	})
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
