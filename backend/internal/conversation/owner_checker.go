package conversation

import (
	"context"

	"github.com/google/uuid"
)

// OwnerChecker adapts conversation membership for entitlement owner checks.
type OwnerChecker struct {
	repo *Repository
}

// NewOwnerChecker creates an owner checker for paid-channel configuration.
func NewOwnerChecker(repo *Repository) *OwnerChecker {
	return &OwnerChecker{repo: repo}
}

// IsChannelOwner returns true when the user owns the channel conversation.
func (c *OwnerChecker) IsChannelOwner(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	member, err := c.repo.GetMember(ctx, channelID, userID)
	if err != nil {
		return false, err
	}
	return member.Role == RoleOwner, nil
}
