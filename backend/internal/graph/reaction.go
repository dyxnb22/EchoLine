package graph

import (
	"context"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/reaction"
)

// ReactionService adapts reaction repository for GraphQL.
type ReactionService struct {
	repo *reaction.Repository
}

// NewReactionService creates a GraphQL reaction adapter.
func NewReactionService(repo *reaction.Repository) *ReactionService {
	return &ReactionService{repo: repo}
}

// Add adds a reaction.
func (s *ReactionService) Add(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	_, err := s.repo.Add(ctx, messageID, userID, emoji)
	return err
}
