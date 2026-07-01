package delivery

import (
	"testing"
)

func TestStatusRankForwardOnly(t *testing.T) {
	if statusRank[StatusRead] <= statusRank[StatusDelivered] {
		t.Fatal("read should outrank delivered")
	}
	if statusRank[StatusDelivered] <= statusRank[StatusSent] {
		t.Fatal("delivered should outrank sent")
	}
}
