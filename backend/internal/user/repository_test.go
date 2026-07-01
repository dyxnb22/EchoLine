package user

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/echoline/echoline/backend/internal/migrate"
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

func TestRepositoryCreateAndGet(t *testing.T) {
	repo := openTestRepo(t)
	ctx := context.Background()

	hash := "argon2id$v=19$m=65536,t=1,p=4$c2FsdHNhbHQ$cGFzc3dvcmQ"

	created, err := repo.Create(ctx, "alice_"+t.Name(), "Alice", hash)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	byUsername, err := repo.GetByUsername(ctx, created.Username)
	if err != nil {
		t.Fatalf("GetByUsername() error = %v", err)
	}
	if byUsername.ID != created.ID {
		t.Fatalf("user id mismatch")
	}

	byID, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if byID.Username != created.Username {
		t.Fatalf("username = %q, want %q", byID.Username, created.Username)
	}
}

func TestRepositoryDuplicateUsername(t *testing.T) {
	repo := openTestRepo(t)
	ctx := context.Background()

	username := "dup_" + t.Name()
	hash := "argon2id$v=19$m=65536,t=1,p=4$c2FsdHNhbHQ$cGFzc3dvcmQ"

	if _, err := repo.Create(ctx, username, "Alice", hash); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err := repo.Create(ctx, username, "Alice 2", hash)
	if !errors.Is(err, ErrDuplicateUsername) {
		t.Fatalf("error = %v, want ErrDuplicateUsername", err)
	}
}
