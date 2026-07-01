package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserSummary is a minimal user record for admin listing.
type UserSummary struct {
	ID          uuid.UUID
	Username    string
	DisplayName string
	IsAdmin     bool
	CreatedAt   time.Time
}

// ReportSummary is a message report for admin review.
type ReportSummary struct {
	ID             uuid.UUID
	ReporterID     uuid.UUID
	MessageID      uuid.UUID
	ConversationID uuid.UUID
	Reason         string
	CreatedAt      time.Time
}

// AuditSummary is an audit log entry.
type AuditSummary struct {
	ID           uuid.UUID
	ActorID      *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   string
	Metadata     map[string]any
	CreatedAt    time.Time
}

// Repository provides admin data access.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates an admin repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ListUsers returns recent users.
func (r *Repository) ListUsers(ctx context.Context, limit int) ([]UserSummary, error) {
	const q = `
		SELECT id, username, display_name, COALESCE(is_admin, false), created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var out []UserSummary
	for rows.Next() {
		var u UserSummary
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.IsAdmin, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// ListReports returns recent message reports.
func (r *Repository) ListReports(ctx context.Context, limit int) ([]ReportSummary, error) {
	const q = `
		SELECT id, reporter_id, message_id, conversation_id, reason, created_at
		FROM message_reports
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	var out []ReportSummary
	for rows.Next() {
		var rp ReportSummary
		if err := rows.Scan(&rp.ID, &rp.ReporterID, &rp.MessageID, &rp.ConversationID, &rp.Reason, &rp.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		out = append(out, rp)
	}
	return out, rows.Err()
}

// ListAuditLogs returns recent audit entries.
func (r *Repository) ListAuditLogs(ctx context.Context, limit int) ([]AuditSummary, error) {
	const q = `
		SELECT id, actor_id, action, resource_type, resource_id, metadata, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list audit: %w", err)
	}
	defer rows.Close()

	var out []AuditSummary
	for rows.Next() {
		var a AuditSummary
		var metaJSON []byte
		if err := rows.Scan(&a.ID, &a.ActorID, &a.Action, &a.ResourceType, &a.ResourceID, &metaJSON, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit: %w", err)
		}
		_ = json.Unmarshal(metaJSON, &a.Metadata)
		out = append(out, a)
	}
	return out, rows.Err()
}
