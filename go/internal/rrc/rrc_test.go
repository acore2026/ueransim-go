package rrc

import (
	"encoding/hex"
	"testing"
)

func TestBuildRRCSetupRequest(t *testing.T) {
	// Sample UE Identity (39 bits)
	ueID := uint64(0x7FFFFFFFFF)
	
	res := BuildRRCSetupRequest(ueID)
	t.Logf("RRCSetupRequest: %s", hex.EncodeToString(res))
	
	if len(res) == 0 {
		t.Error("result is empty")
	}
}
