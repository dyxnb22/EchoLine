package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository appends audit log entries.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates an audit repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Entry is a single audit record.
type Entry struct {
	ID           uuid.UUID
	ActorID      *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   string
	Metadata     map[string]any
	CreatedAt    time.Time
}

// Append inserts an audit log row.
func (r *Repository) Append(ctx context.Context, actorID *uuid.UUID, action, resourceType, resourceID string, metadata map[string]any) error {
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	const q = `
		INSERT INTO audit_logs (id, actor_id, action, resource_type, resource_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.pool.Exec(ctx, q, uuid.New(), actorID, action, resourceType, resourceID, metaJSON, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("append audit: %w", err)
	}
	return nil
}

// LogLogin records a login attempt.
func (r *Repository) LogLogin(ctx context.Context, userID *uuid.UUID, username string, success bool, ip string) error {
	if r == nil {
		return nil
	}
	meta := map[string]any{
		"username": username,
		"success":  success,
		"ip":       ip,
	}
	actor := userID
	return r.Append(ctx, actor, "auth.login", "user", username, meta)
}
