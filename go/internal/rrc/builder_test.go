package rrc

import (
	"encoding/hex"
	"testing"
)

func TestBuildRRCSetupComplete(t *testing.T) {
	nasPdu := []byte{0x7E, 0x00, 0x41} // Registration Request
	rrcPdu := BuildRRCSetupComplete(nasPdu)
	
	// Bits:
	// 0 (1b) | 0010 (4b) | 00 (2b) | 0 (1b) -> 00010000 -> 0x10
	
	t.Logf("RRC PDU: %s", hex.EncodeToString(rrcPdu))
	
	if len(rrcPdu) < 3 {
		t.Fatalf("RRC PDU too short: %d", len(rrcPdu))
	}
	
	if rrcPdu[0] != 0x10 {
		t.Errorf("expected pdu[0]=0x10, got 0x%02x", rrcPdu[0])
	}
}

func TestBuildDLInformationTransfer(t *testing.T) {
	nasPdu := []byte{0x7E, 0x00, 0x56} // Authentication Request
	rrcPdu := BuildDLInformationTransfer(nasPdu)
	
	// Bits:
	// 0 (1b) | 0101 (4b index 5) | 00 (2b trans) | 0 (1b critical) -> 00101000 -> 0x28
	
	t.Logf("RRC PDU: %s", hex.EncodeToString(rrcPdu))
	
	if len(rrcPdu) < 3 {
		t.Fatalf("RRC PDU too short: %d", len(rrcPdu))
	}
	
	if rrcPdu[0] != 0x28 {
		t.Errorf("expected pdu[0]=0x28, got 0x%02x", rrcPdu[0])
	}
}
