package admin

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CompositeAdminChecker grants admin when env list OR users.is_admin is true.
type CompositeAdminChecker struct {
	static *StaticAdminChecker
	pool   *pgxpool.Pool
}

// NewCompositeAdminChecker combines env UUIDs with DB is_admin flag.
func NewCompositeAdminChecker(static *StaticAdminChecker, pool *pgxpool.Pool) *CompositeAdminChecker {
	return &CompositeAdminChecker{static: static, pool: pool}
}

// IsAdmin returns true for configured admins or DB is_admin users.
func (c *CompositeAdminChecker) IsAdmin(userID uuid.UUID) bool {
	if c.static != nil && c.static.IsAdmin(userID) {
		return true
	}
	if c.pool == nil {
		return false
	}
	var isAdmin bool
	err := c.pool.QueryRow(context.Background(),
		`SELECT is_admin FROM users WHERE id = $1`, userID).Scan(&isAdmin)
	return err == nil && isAdmin
}
