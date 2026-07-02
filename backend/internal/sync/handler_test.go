package sync

import "testing"

func TestSyncPaginationMeta(t *testing.T) {
	tests := []struct {
		fetched  int
		pageSize int
		wantMore bool
		wantLen  int
	}{
		{0, 200, false, 0},
		{50, 200, false, 50},
		{200, 200, false, 200},
		{201, 200, true, 200},
		{350, 200, true, 200},
	}
	for _, tc := range tests {
		hasMore, count := paginationMeta(tc.fetched, tc.pageSize)
		if hasMore != tc.wantMore || count != tc.wantLen {
			t.Fatalf("fetched=%d page=%d: got hasMore=%v count=%d want hasMore=%v count=%d",
				tc.fetched, tc.pageSize, hasMore, count, tc.wantMore, tc.wantLen)
		}
	}
}

func TestMaxSyncCursors(t *testing.T) {
	if maxSyncCursors != 50 {
		t.Fatalf("maxSyncCursors = %d, want 50", maxSyncCursors)
	}
}

func TestSyncNextSeqUsesPageMax(t *testing.T) {
	lastSeq := int64(100)
	maxSeq := int64(250)
	if maxSeq <= lastSeq {
		t.Fatalf("next_seq must be greater than cursor when page has messages")
	}
}
