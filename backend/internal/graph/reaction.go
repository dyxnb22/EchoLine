package graph

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/message"
	"github.com/echoline/echoline/backend/internal/reaction"
)

// ReactionService adapts reaction repository for GraphQL.
type ReactionService struct {
	repo          *reaction.Repository
	conversations *conversation.Repository
	messages      *message.Repository
}

// NewReactionService creates a GraphQL reaction adapter.
func NewReactionService(repo *reaction.Repository, conversations *conversation.Repository, messages *message.Repository) *ReactionService {
	return &ReactionService{repo: repo, conversations: conversations, messages: messages}
}

// Add adds a reaction when the user is a member of the message conversation.
func (s *ReactionService) Add(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	convID, err := s.messages.GetConversationID(ctx, messageID)
	if err != nil {
		if errors.Is(err, message.ErrNotFound) {
			return message.ErrNotFound
		}
		return err
	}
	member, err := s.conversations.IsMember(ctx, convID, userID)
	if err != nil || !member {
		return conversation.ErrForbidden
	}
	_, _, err = s.repo.Add(ctx, messageID, userID, emoji)
	return err
}
