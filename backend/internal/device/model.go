package device

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("device not found")

// Device represents a user login device.
type Device struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	DeviceName string
	Platform   string
	LastSeenAt *time.Time
	CreatedAt  time.Time
}
