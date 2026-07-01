package reliability

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/risk"
)

// D009: fault-injection style unit checks for reliability primitives.
func TestSpamCheckerBlocksRapidDuplicates(t *testing.T) {
	checker := risk.NewSpamChecker()
	user := uuid.New()
	body := "same message body"

	for i := 0; i < 3; i++ {
		if err := checker.CheckDuplicateBody(user, body); err != nil {
			t.Fatalf("send %d should pass: %v", i+1, err)
		}
	}
	if err := checker.CheckDuplicateBody(user, body); err == nil {
		t.Fatal("duplicate beyond threshold should fail")
	}
}

func TestSpamCheckerAllowsDifferentBodies(t *testing.T) {
	checker := risk.NewSpamChecker()
	user := uuid.New()
	if err := checker.CheckDuplicateBody(user, "a"); err != nil {
		t.Fatal(err)
	}
	if err := checker.CheckDuplicateBody(user, "b"); err != nil {
		t.Fatal(err)
	}
	_ = time.Second
}
