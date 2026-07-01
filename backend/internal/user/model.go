package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("user not found")
var ErrDuplicateUsername = errors.New("username already exists")

// User represents a registered account.
type User struct {
	ID           uuid.UUID
	Username     string
	DisplayName  string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
