package tasks

import (
	"testing"
	"time"
)

func TestDuplicateDownlinkSuppression(t *testing.T) {
	h := &GnbGtpTaskHandler{
		lastDl: make(map[uint8]recentDownlink),
	}

	payload := []byte{0x45, 0x00, 0x00, 0x54}
	if h.isDuplicateDownlink(1, 7, payload) {
		t.Fatalf("first packet should not be treated as duplicate")
	}
	if !h.isDuplicateDownlink(1, 7, payload) {
		t.Fatalf("immediate identical packet should be treated as duplicate")
	}

	time.Sleep(25 * time.Millisecond)
	if h.isDuplicateDownlink(1, 7, payload) {
		t.Fatalf("same packet after dedup window should not be treated as duplicate")
	}

	if h.isDuplicateDownlink(1, 8, payload) {
		t.Fatalf("different TEID should not be treated as duplicate")
	}
	if h.isDuplicateDownlink(1, 8, []byte{0x45, 0x00, 0x00, 0x55}) {
		t.Fatalf("different payload should not be treated as duplicate")
	}
}
