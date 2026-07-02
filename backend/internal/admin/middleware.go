package admin

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
)

// AdminChecker verifies admin privileges.
type AdminChecker interface {
	IsAdmin(userID uuid.UUID) bool
}

// StaticAdminChecker grants admin to configured user IDs.
type StaticAdminChecker struct {
	ids map[uuid.UUID]struct{}
}

// NewStaticAdminChecker parses comma-separated UUIDs.
func NewStaticAdminChecker(raw string) *StaticAdminChecker {
	c := &StaticAdminChecker{ids: make(map[uuid.UUID]struct{})}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if id, err := uuid.Parse(part); err == nil {
			c.ids[id] = struct{}{}
		}
	}
	return c
}

// IsAdmin returns true when userID is in the configured admin set.
func (c *StaticAdminChecker) IsAdmin(userID uuid.UUID) bool {
	if c == nil || len(c.ids) == 0 {
		return false
	}
	_, ok := c.ids[userID]
	return ok
}

// RequireAdmin wraps a handler with admin authorization.
func RequireAdmin(authSvc *auth.Service, checker AdminChecker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok {
			apierror.Write(w, r, http.StatusUnauthorized, "unauthorized", "missing auth")
			return
		}
		if checker == nil || !checker.IsAdmin(claims.UserID) {
			apierror.Write(w, r, http.StatusForbidden, "forbidden", "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
