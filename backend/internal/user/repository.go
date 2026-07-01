package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists users in PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a user repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new user.
func (r *Repository) Create(ctx context.Context, username, displayName, passwordHash string) (*User, error) {
	username = strings.TrimSpace(username)
	displayName = strings.TrimSpace(displayName)

	now := time.Now().UTC()
	id := uuid.New()

	const q = `
		INSERT INTO users (id, username, display_name, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, username, display_name, password_hash, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, q, id, username, displayName, passwordHash, now, now)
	u, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateUsername
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

// GetByUsername loads a user by username.
func (r *Repository) GetByUsername(ctx context.Context, username string) (*User, error) {
	const q = `
		SELECT id, username, display_name, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	row := r.pool.QueryRow(ctx, q, strings.TrimSpace(username))
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

// GetByID loads a user by ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	const q = `
		SELECT id, username, display_name, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, q, id)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanUser(row scannable) (*User, error) {
	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.DisplayName, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}
