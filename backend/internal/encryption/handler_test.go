package encryption

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestBundlePayload(t *testing.T) {
	kb := &KeyBundle{
		ID:        uuid.New(),
		DeviceID:  "dev1",
		PublicKey: "pk",
	}
	p := bundlePayload(kb)
	if p["device_id"] != "dev1" {
		t.Fatal("unexpected payload")
	}
}

func TestRepositoryNilPool(t *testing.T) {
	// compile-time check that handler constructs
	h := NewHandler(nil)
	if h == nil {
		t.Fatal("nil handler")
	}
	_ = context.Background()
}
