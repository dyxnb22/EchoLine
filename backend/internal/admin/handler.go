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
}

// NewHandler creates an admin handler.
func NewHandler(pool *pgxpool.Pool, authSvc *auth.Service) *Handler {
	return &Handler{pool: pool, auth: authSvc}
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
