package message

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/media"
	"github.com/echoline/echoline/backend/internal/risk"
)

// Broadcaster pushes realtime events to online users.
type Broadcaster interface {
	BroadcastMessageCreated(ctx context.Context, convID uuid.UUID, msg *Message, excludeSender bool, senderID uuid.UUID) error
}

// Service coordinates message persistence and realtime fanout.
type Service struct {
	repo          *Repository
	conversations *conversation.Repository
	attachments   *media.Repository
	broadcaster   Broadcaster
	spamChecker   *risk.SpamChecker
}

// SetBroadcaster attaches a realtime broadcaster after construction.
func (s *Service) SetBroadcaster(b Broadcaster) {
	s.broadcaster = b
}

// NewService creates a message service.
func NewService(repo *Repository, conversations *conversation.Repository, attachments *media.Repository, broadcaster Broadcaster) *Service {
	return &Service{
		repo:          repo,
		conversations: conversations,
		attachments:   attachments,
		broadcaster:   broadcaster,
		spamChecker:   risk.NewSpamChecker(),
	}
}

// SendInput carries optional attachment metadata for non-text messages.
type SendInput struct {
	ClientMsgID string
	Type        Type
	Body        string
	ObjectKey   string
}

// Send persists a message and notifies online members.
func (s *Service) Send(ctx context.Context, convID, senderID uuid.UUID, input SendInput) (*Message, error) {
	if err := s.conversations.CanPublish(ctx, convID, senderID); err != nil {
		return nil, err
	}

	if input.Body != "" && s.spamChecker != nil {
		if err := s.spamChecker.CheckDuplicateBody(senderID, input.Body); err != nil {
			if errors.Is(err, risk.ErrSpamDetected) {
				return nil, errors.New("rate limit: duplicate message body")
			}
		}
	}

	msgType := input.Type
	if msgType == "" {
		msgType = TypeText
	}

	var attachmentID *uuid.UUID
	if input.ObjectKey != "" {
		if s.attachments == nil {
			return nil, errors.New("attachments not configured")
		}
		att, err := s.attachments.GetUnlinkedByObjectKey(ctx, senderID, input.ObjectKey)
		if err != nil {
			return nil, err
		}
		attachmentID = &att.ID
		if msgType == TypeText {
			if strings.HasPrefix(att.MimeType, "image/") {
				msgType = TypeImage
			} else {
				msgType = TypeFile
			}
		}
	}

	msg, err := s.repo.Create(ctx, convID, senderID, input.ClientMsgID, msgType, input.Body, attachmentID)
	if err != nil {
		return nil, err
	}

	if s.broadcaster != nil {
		_ = s.broadcaster.BroadcastMessageCreated(ctx, convID, msg, true, senderID)
	}
	return msg, nil
}

// ListSince returns messages with seq greater than afterSeq.
func (s *Service) ListSince(ctx context.Context, convID uuid.UUID, afterSeq int64, limit int) ([]Message, error) {
	return s.repo.ListSince(ctx, convID, afterSeq, limit)
}

// List returns paginated history.
func (s *Service) List(ctx context.Context, convID uuid.UUID, beforeSeq *int64, limit int) ([]Message, error) {
	return s.repo.List(ctx, convID, beforeSeq, limit)
}

// Edit updates a message body for the sender.
func (s *Service) Edit(ctx context.Context, convID, messageID, senderID uuid.UUID, body string) (*Message, error) {
	if err := s.conversations.CanPublish(ctx, convID, senderID); err != nil {
		return nil, err
	}
	return s.repo.Edit(ctx, convID, messageID, senderID, body)
}

// Recall marks a message recalled; sender or admin may recall.
func (s *Service) Recall(ctx context.Context, convID, messageID, actorID uuid.UUID) (*Message, error) {
	msg, err := s.repo.GetByID(ctx, convID, messageID)
	if err != nil {
		return nil, err
	}
	if msg.SenderID != actorID {
		member, err := s.conversations.GetMember(ctx, convID, actorID)
		if err != nil || !conversation.CanManageMembers(member.Role) {
			return nil, conversation.ErrForbidden
		}
	}
	return s.repo.Recall(ctx, convID, messageID)
}

// ToCreatedPayload converts a message for WS/REST responses.
func ToCreatedPayload(msg *Message) map[string]any {
	return ToCreatedPayloadWithAttachment(msg, nil)
}

// ToCreatedPayloadWithAttachment converts a message and optional attachment metadata.
func ToCreatedPayloadWithAttachment(msg *Message, att *media.Attachment) map[string]any {
	payload := map[string]any{
		"id":              msg.ID,
		"conversation_id": msg.ConversationID,
		"sender_id":       msg.SenderID,
		"client_msg_id":   msg.ClientMsgID,
		"seq":             msg.Seq,
		"type":            msg.Type,
		"body":            msg.Body,
		"status":          msg.Status,
		"created_at":      msg.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":      msg.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if att != nil {
		payload["attachment"] = map[string]any{
			"id":         att.ID,
			"object_key": att.ObjectKey,
			"mime_type":  att.MimeType,
			"size_bytes": att.SizeBytes,
		}
	}
	return payload
}
