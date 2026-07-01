package risk

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ErrSpamDetected is returned when a duplicate message body is detected.
var ErrSpamDetected = errors.New("spam: duplicate message body")

const (
	defaultWindow    = 10 * time.Second
	defaultMaxRepeat = 3
)

type entry struct {
	count     int
	firstSeen time.Time
}

// SpamChecker is a simple in-memory duplicate body detector.
type SpamChecker struct {
	mu      sync.Mutex
	records map[string]*entry
	window  time.Duration
	maxRep  int
}

// NewSpamChecker creates a spam checker with default thresholds.
func NewSpamChecker() *SpamChecker {
	s := &SpamChecker{
		records: make(map[string]*entry),
		window:  defaultWindow,
		maxRep:  defaultMaxRepeat,
	}
	go s.gc()
	return s
}

// CheckDuplicateBody returns ErrSpamDetected if userID sent body more than maxRepeat times within window.
func (s *SpamChecker) CheckDuplicateBody(userID uuid.UUID, body string) error {
	key := s.key(userID, body)
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.records[key]
	if !ok || now.Sub(e.firstSeen) > s.window {
		s.records[key] = &entry{count: 1, firstSeen: now}
		return nil
	}
	e.count++
	if e.count > s.maxRep {
		return ErrSpamDetected
	}
	return nil
}

func (s *SpamChecker) key(userID uuid.UUID, body string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", userID.String(), body)))
	return hex.EncodeToString(h[:])
}

// gc periodically removes stale entries.
func (s *SpamChecker) gc() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for k, e := range s.records {
			if now.Sub(e.firstSeen) > s.window*2 {
				delete(s.records, k)
			}
		}
		s.mu.Unlock()
	}
}
