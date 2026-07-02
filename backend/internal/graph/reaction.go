package graph

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/message"
	"github.com/echoline/echoline/backend/internal/reaction"
)

type conversationMemberChecker interface {
	IsMember(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
}

type messageConversationResolver interface {
	GetConversationID(ctx context.Context, messageID uuid.UUID) (uuid.UUID, error)
}

// ReactionService adapts reaction repository for GraphQL with membership checks.
type ReactionService struct {
	repo          *reaction.Repository
	conversations conversationMemberChecker
	messages      messageConversationResolver
}

// NewReactionService creates a GraphQL reaction adapter.
func NewReactionService(
	repo *reaction.Repository,
	conversations conversationMemberChecker,
	messages messageConversationResolver,
) *ReactionService {
	return &ReactionService{repo: repo, conversations: conversations, messages: messages}
}

// Add adds a reaction after verifying conversation membership.
func (s *ReactionService) Add(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	convID, err := s.messages.GetConversationID(ctx, messageID)
	if err != nil {
		if errors.Is(err, message.ErrNotFound) {
			return fmt.Errorf("message not found")
		}
		return fmt.Errorf("failed to resolve message")
	}
	member, err := s.conversations.IsMember(ctx, convID, userID)
	if err != nil {
		return fmt.Errorf("failed to check membership")
	}
	if !member {
		return fmt.Errorf("not a conversation member")
	}
	_, _, err = s.repo.Add(ctx, messageID, userID, emoji)
	return err
}
