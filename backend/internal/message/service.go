package message

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/media"
	"github.com/echoline/echoline/backend/internal/middleware"
	"github.com/echoline/echoline/backend/internal/risk"
	"github.com/echoline/echoline/backend/internal/validate"
)

// Broadcaster pushes realtime events to online users.
type Broadcaster interface {
	BroadcastMessageCreated(ctx context.Context, convID uuid.UUID, msg *Message, excludeSender bool, senderID uuid.UUID) error
	BroadcastMessageEdited(ctx context.Context, convID uuid.UUID, msg *Message) error
	BroadcastMessageRecalled(ctx context.Context, convID uuid.UUID, msg *Message) error
}

// BlockChecker checks whether one user has blocked another.
type BlockChecker interface {
	IsBlocked(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error)
}

// Service coordinates message persistence and realtime fanout.
type Service struct {
	repo          *Repository
	conversations *conversation.Repository
	attachments   *media.Repository
	broadcaster   Broadcaster
	spamChecker   *risk.SpamChecker
	blockChecker  BlockChecker
}

// SetBroadcaster attaches a realtime broadcaster after construction.
func (s *Service) SetBroadcaster(b Broadcaster) {
	s.broadcaster = b
}

// SetBlockChecker attaches a block checker for DM send validation.
func (s *Service) SetBlockChecker(bc BlockChecker) {
	s.blockChecker = bc
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

// ErrBlocked is returned when the recipient has blocked the sender.
var ErrBlocked = fmt.Errorf("recipient has blocked you")

// SendReply persists a threaded reply through the main message pipeline.
func (s *Service) SendReply(ctx context.Context, convID, senderID, parentMsgID uuid.UUID, body string) (*Message, error) {
	if err := s.conversations.CanPublish(ctx, convID, senderID); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetByID(ctx, convID, parentMsgID); err != nil {
		return nil, err
	}
	body = middleware.SanitizeBody(body)
	if err := validate.MessageBody(body, false); err != nil {
		return nil, err
	}
	parent := parentMsgID
	msg, err := s.repo.Create(ctx, convID, senderID, uuid.New().String(), TypeText, body, nil, CreateOptions{ParentMessageID: &parent})
	if err != nil {
		return nil, err
	}
	if s.broadcaster != nil {
		_ = s.broadcaster.BroadcastMessageCreated(ctx, convID, msg, true, senderID)
	}
	return msg, nil
}

// Forward copies a message body into another conversation via the main pipeline.
func (s *Service) Forward(ctx context.Context, sourceMsgID, targetConvID, senderID uuid.UUID) (*Message, error) {
	src, err := s.repo.GetByMessageID(ctx, sourceMsgID)
	if err != nil {
		return nil, err
	}
	member, err := s.conversations.IsMember(ctx, src.ConversationID, senderID)
	if err != nil || !member {
		return nil, conversation.ErrForbidden
	}
	return s.Send(ctx, targetConvID, senderID, SendInput{
		ClientMsgID: uuid.New().String(),
		Type:        src.Type,
		Body:        src.Body,
	})
}

// Send persists a message and notifies online members.
func (s *Service) Send(ctx context.Context, convID, senderID uuid.UUID, input SendInput) (*Message, error) {
	if err := s.conversations.CanPublish(ctx, convID, senderID); err != nil {
		return nil, err
	}

	// Block check: if this is a direct conversation, verify the peer has not blocked the sender.
	if s.blockChecker != nil {
		peerID, err := s.conversations.GetDirectPeer(ctx, convID, senderID)
		if err == nil && peerID != uuid.Nil {
			blocked, err := s.blockChecker.IsBlocked(ctx, peerID, senderID)
			if err == nil && blocked {
				return nil, ErrBlocked
			}
		}
	}

	// Strip HTML from message body before persistence.
	input.Body = middleware.SanitizeBody(input.Body)

	if err := validate.MessageBody(input.Body, input.ObjectKey != ""); err != nil {
		return nil, err
	}

	if input.Body != "" && s.spamChecker != nil {
		if err := s.spamChecker.CheckDuplicateBody(senderID, input.Body); err != nil {
			if errors.Is(err, risk.ErrSpamDetected) {
				return nil, errors.New("rate limit: duplicate message body")
			}
		}
	}

	clientMsgID, err := validate.ClientMsgID(input.ClientMsgID)
	if err != nil {
		return nil, err
	}
	input.ClientMsgID = clientMsgID

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
	body = middleware.SanitizeBody(body)
	if err := validate.MessageBody(body, false); err != nil {
		return nil, err
	}
	msg, err := s.repo.Edit(ctx, convID, messageID, senderID, body)
	if err != nil {
		return nil, err
	}
	if s.broadcaster != nil {
		_ = s.broadcaster.BroadcastMessageEdited(ctx, convID, msg)
	}
	return msg, nil
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
	recalled, err := s.repo.Recall(ctx, convID, messageID)
	if err != nil {
		return nil, err
	}
	if s.broadcaster != nil {
		_ = s.broadcaster.BroadcastMessageRecalled(ctx, convID, recalled)
	}
	return recalled, nil
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
