package entitlement

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestErrEntitlementRequired(t *testing.T) {
	if ErrEntitlementRequired.Error() == "" {
		t.Fatal("expected message")
	}
	_ = uuid.New()
	_ = context.Background()
}
