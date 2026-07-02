package outbox

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/echoline/echoline/backend/internal/migrate"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func openTestRepo(t *testing.T) *Repository {
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

	return NewRepository(pool)
}

func TestEventFields(t *testing.T) {
	e := Event{
		ID:        uuid.New(),
		Topic:     "message.created",
		Payload:   []byte(`{"id":"m1"}`),
		Attempts:  0,
		CreatedAt: time.Now().UTC(),
	}
	if e.Topic == "" {
		t.Fatal("expected topic")
	}
}

func TestRepositoryReclaimStaleProcessing(t *testing.T) {
	repo := openTestRepo(t)
	ctx := context.Background()

	id := uuid.New()
	staleAt := time.Now().UTC().Add(-10 * time.Minute)
	_, err := repo.pool.Exec(ctx, `
		INSERT INTO outbox_events (id, topic, payload, status, attempts, created_at, processing_started_at)
		VALUES ($1, 'message.created', '{"id":"m1"}'::jsonb, 'processing', 0, $2, $3)
	`, id, staleAt, staleAt)
	if err != nil {
		t.Fatalf("insert processing row: %v", err)
	}

	freshID := uuid.New()
	freshAt := time.Now().UTC()
	_, err = repo.pool.Exec(ctx, `
		INSERT INTO outbox_events (id, topic, payload, status, attempts, created_at, processing_started_at)
		VALUES ($1, 'message.created', '{"id":"m2"}'::jsonb, 'processing', 0, $2, $3)
	`, freshID, freshAt, freshAt)
	if err != nil {
		t.Fatalf("insert fresh processing row: %v", err)
	}

	reclaimed, err := repo.ReclaimStaleProcessing(ctx, time.Now().UTC().Add(-5*time.Minute))
	if err != nil {
		t.Fatalf("ReclaimStaleProcessing() error = %v", err)
	}
	if reclaimed != 1 {
		t.Fatalf("reclaimed = %d, want 1", reclaimed)
	}

	var staleStatus string
	if err := repo.pool.QueryRow(ctx, `SELECT status FROM outbox_events WHERE id = $1`, id).Scan(&staleStatus); err != nil {
		t.Fatalf("query stale row: %v", err)
	}
	if staleStatus != "pending" {
		t.Fatalf("stale status = %q, want pending", staleStatus)
	}

	var freshStatus string
	if err := repo.pool.QueryRow(ctx, `SELECT status FROM outbox_events WHERE id = $1`, freshID).Scan(&freshStatus); err != nil {
		t.Fatalf("query fresh row: %v", err)
	}
	if freshStatus != "processing" {
		t.Fatalf("fresh status = %q, want processing", freshStatus)
	}
}

func TestRepositoryFetchPendingSetsProcessingStartedAt(t *testing.T) {
	repo := openTestRepo(t)
	ctx := context.Background()

	if err := repo.Enqueue(ctx, "message.created", []byte(`{"id":"m3"}`)); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	before := time.Now().UTC()
	events, err := repo.FetchPending(ctx, 1)
	if err != nil {
		t.Fatalf("FetchPending() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}

	var startedAt time.Time
	if err := repo.pool.QueryRow(ctx, `
		SELECT processing_started_at FROM outbox_events WHERE id = $1
	`, events[0].ID).Scan(&startedAt); err != nil {
		t.Fatalf("query processing_started_at: %v", err)
	}
	if startedAt.Before(before) || startedAt.After(time.Now().UTC().Add(time.Second)) {
		t.Fatalf("processing_started_at = %v, want near now", startedAt)
	}
}
