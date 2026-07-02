package payment

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type mockPaidChannels struct {
	required bool
}

func (m *mockPaidChannels) RequiresEntitlement(_ context.Context, _ uuid.UUID) (bool, error) {
	return m.required, nil
}

func TestValidateChannelPaymentRefRequiresPaidChannel(t *testing.T) {
	h := &Handler{paidChannels: &mockPaidChannels{required: false}}
	channelID := uuid.New()
	_, err := h.validateChannelPaymentRef(context.Background(), "channel:"+channelID.String())
	if err == nil {
		t.Fatal("expected error when channel does not require payment")
	}
}

func TestValidateChannelPaymentRefAllowsPaidChannel(t *testing.T) {
	h := &Handler{paidChannels: &mockPaidChannels{required: true}}
	channelID := uuid.New()
	got, err := h.validateChannelPaymentRef(context.Background(), "channel:"+channelID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != channelID {
		t.Fatalf("channelID = %v, want %v", got, channelID)
	}
}
