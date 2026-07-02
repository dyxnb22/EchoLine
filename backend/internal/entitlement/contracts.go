package entitlement

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrEntitlementRequired is returned when a paid channel lacks an active grant.
var ErrEntitlementRequired = errors.New("channel requires paid entitlement")

// ErrNotPaidChannel is returned when the conversation is not a channel.
var ErrNotPaidChannel = errors.New("conversation is not a paid channel")

// Gate checks whether a user may subscribe to a channel.
type Gate interface {
	CanSubscribe(ctx context.Context, userID, channelID uuid.UUID) error
}

// OwnerChecker verifies channel ownership for paid-channel configuration.
type OwnerChecker interface {
	IsChannelOwner(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
}
