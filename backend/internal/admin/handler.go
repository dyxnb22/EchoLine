package admin

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// Handler exposes admin endpoints.
type Handler struct {
	pool *pgxpool.Pool
	auth *auth.Service
	repo *Repository
}

// NewHandler creates an admin handler.
func NewHandler(pool *pgxpool.Pool, authSvc *auth.Service) *Handler {
	return &Handler{pool: pool, auth: authSvc, repo: NewRepository(pool)}
}

// HandleHealth is a deep health check with DB ping.
// GET /api/admin/health
// Auth stub: requires valid JWT (any user). Production would enforce admin role.
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
		return
	}

	status := "ok"
	dbStatus := "ok"
	code := http.StatusOK

	if h.pool != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := h.pool.Ping(ctx); err != nil {
			dbStatus = "error"
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
	}

	apierror.WriteJSON(w, code, map[string]any{
		"status":    status,
		"db":        dbStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// HandleListUsers lists users for admin review.
// GET /api/admin/users
func (h *Handler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.ListUsers(r.Context(), 100)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list users")
		return
	}
	items := make([]map[string]any, 0, len(users))
	for _, u := range users {
		items = append(items, map[string]any{
			"id":           u.ID,
			"username":     u.Username,
			"display_name": u.DisplayName,
			"is_admin":     u.IsAdmin,
			"created_at":   u.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	apierror.WriteJSON(w, http.StatusOK, map[string]any{"users": items})
}

// HandleListReports lists message reports.
// GET /api/admin/reports
func (h *Handler) HandleListReports(w http.ResponseWriter, r *http.Request) {
	reports, err := h.repo.ListReports(r.Context(), 100)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list reports")
		return
	}
	items := make([]map[string]any, 0, len(reports))
	for _, rp := range reports {
		items = append(items, map[string]any{
			"id":              rp.ID,
			"reporter_id":     rp.ReporterID,
			"message_id":      rp.MessageID,
			"conversation_id": rp.ConversationID,
			"reason":          rp.Reason,
			"created_at":      rp.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	apierror.WriteJSON(w, http.StatusOK, map[string]any{"reports": items})
}

// HandleListAuditLogs lists audit log entries.
// GET /api/admin/audit-logs
func (h *Handler) HandleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	entries, err := h.repo.ListAuditLogs(r.Context(), 100)
	if err != nil {
		apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "failed to list audit logs")
		return
	}
	items := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		items = append(items, map[string]any{
			"id":            e.ID,
			"actor_id":      e.ActorID,
			"action":        e.Action,
			"resource_type": e.ResourceType,
			"resource_id":   e.ResourceID,
			"metadata":      e.Metadata,
			"created_at":    e.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	apierror.WriteJSON(w, http.StatusOK, map[string]any{"entries": items})
}
