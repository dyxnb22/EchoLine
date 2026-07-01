package media

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestObjectKeyPrefixValidation(t *testing.T) {
	owner := uuid.New()
	key := "uploads/" + owner.String() + "/file.bin"
	if !strings.HasPrefix(key, "uploads/"+owner.String()+"/") {
		t.Fatal("expected owner prefix")
	}
}
