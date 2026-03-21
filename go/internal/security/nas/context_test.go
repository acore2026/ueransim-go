package nas

import (
	"testing"
)

func TestNasCountRollover(t *testing.T) {
	sc := NewSecurityContext(nil, nil, 0, 0)

	// Initial state
	if sc.DlCount.Uint32() != 0 {
		t.Errorf("expected 0, got %d", sc.DlCount.Uint32())
	}

	// Normal increment
	sc.UpdateDlCount(sc.EstimatedDlCount(10))
	if sc.DlCount.Uint32() != 10 {
		t.Errorf("expected 10, got %d", sc.DlCount.Uint32())
	}

	// Rollover
	sc.UpdateDlCount(sc.EstimatedDlCount(255))
	if sc.DlCount.Uint32() != 255 {
		t.Errorf("expected 255, got %d", sc.DlCount.Uint32())
	}

	sc.UpdateDlCount(sc.EstimatedDlCount(0))
	if sc.DlCount.Uint32() != 256 {
		t.Errorf("expected 256, got %d", sc.DlCount.Uint32())
	}
	if sc.DlCount.Overflow != 1 {
		t.Errorf("expected overflow 1, got %d", sc.DlCount.Overflow)
	}

	// Multiple rollovers
	sc.DlCount.Overflow = 0xFFFFFE
	sc.DlCount.SQN = 254
	sc.UpdateDlCount(sc.EstimatedDlCount(255)) // No rollover
	if sc.DlCount.Overflow != 0xFFFFFE {
		t.Errorf("expected overflow 0xFFFFFE, got %x", sc.DlCount.Overflow)
	}
	sc.UpdateDlCount(sc.EstimatedDlCount(0)) // Rollover: 255 > 0
	if sc.DlCount.Overflow != 0xFFFFFF {
		t.Errorf("expected overflow 0xFFFFFF, got %x", sc.DlCount.Overflow)
	}
	sc.UpdateDlCount(sc.EstimatedDlCount(0)) // No change since SQN is already 0
	if sc.DlCount.Overflow != 0xFFFFFF {
		t.Errorf("expected no overflow change, got %x", sc.DlCount.Overflow)
	}
	sc.DlCount.SQN = 255
	sc.UpdateDlCount(sc.EstimatedDlCount(0)) // Rollover: 255 > 0 -> Wrap to 0
	if sc.DlCount.Overflow != 0x000000 {
		t.Errorf("expected overflow wrap to 0, got %x", sc.DlCount.Overflow)
	}
}

func TestReplayProtection(t *testing.T) {
	sc := NewSecurityContext(nil, nil, 0, 0)

	if !sc.CheckForReplay(1) {
		t.Error("expected true for first SN 1")
	}
	if sc.CheckForReplay(1) {
		t.Error("expected false for duplicate SN 1")
	}

	for i := 2; i <= 17; i++ {
		if !sc.CheckForReplay(uint8(i)) {
			t.Errorf("expected true for SN %d", i)
		}
	}

	// 1 should have been evicted from the window (size 16)
	// Window now contains 2, 3, ..., 17
	if !sc.CheckForReplay(1) {
		t.Error("expected true for SN 1 after eviction")
	}
}
