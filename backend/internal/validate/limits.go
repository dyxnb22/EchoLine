package validate

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Field length limits aligned with DB columns and API docs.
const (
	MaxUsernameLen    = 64
	MaxDisplayNameLen = 128
	MaxMessageBodyLen = 65535
	MinPasswordLen    = 8
)

var (
	ErrUsernameEmpty    = errors.New("username is required")
	ErrUsernameTooLong  = fmt.Errorf("username exceeds %d characters", MaxUsernameLen)
	ErrDisplayNameLong  = fmt.Errorf("display_name exceeds %d characters", MaxDisplayNameLen)
	ErrPasswordShort    = fmt.Errorf("password must be at least %d characters", MinPasswordLen)
	ErrMessageBodyLong  = fmt.Errorf("message body exceeds %d characters", MaxMessageBodyLen)
	ErrMessageBodyEmpty = errors.New("message body or attachment is required")
	ErrClientMsgIDEmpty = errors.New("client_msg_id is required")
	ErrClientMsgIDInvalid = errors.New("client_msg_id must be a valid UUID")
)

// Username validates registration/login username.
func Username(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", ErrUsernameEmpty
	}
	if utf8.RuneCountInString(s) > MaxUsernameLen {
		return "", ErrUsernameTooLong
	}
	return s, nil
}

// DisplayName validates optional display name.
func DisplayName(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if utf8.RuneCountInString(s) > MaxDisplayNameLen {
		return "", ErrDisplayNameLong
	}
	return s, nil
}

// Password validates minimum password length.
func Password(raw string) error {
	if utf8.RuneCountInString(raw) < MinPasswordLen {
		return ErrPasswordShort
	}
	return nil
}

// ClientMsgID validates a non-empty UUID client message id.
func ClientMsgID(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", ErrClientMsgIDEmpty
	}
	if _, err := uuid.Parse(s); err != nil {
		return "", ErrClientMsgIDInvalid
	}
	return s, nil
}

// MessageBody validates message text length (after sanitization).
func MessageBody(body string, hasAttachment bool) error {
	if body == "" && !hasAttachment {
		return ErrMessageBodyEmpty
	}
	if utf8.RuneCountInString(body) > MaxMessageBodyLen {
		return ErrMessageBodyLong
	}
	return nil
}
