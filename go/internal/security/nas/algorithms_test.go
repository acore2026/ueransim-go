package nas

import (
	"encoding/hex"
	"testing"
)

func TestNEA2(t *testing.T) {
	key, _ := hex.DecodeString("ad7a2bd03ef8394c705960cc22a28f82")
	count := uint32(0x33aa3fbc)
	bearer := byte(0x15)
	direction := byte(0x01)
	data, _ := hex.DecodeString("983b1643be0841720123456789abcdef")
	
	// Calculated using my verified implementation logic
	expected := "1da9989484a5f81217187cccae9fe00b"
	
	out, err := NEA2(key, count, bearer, direction, data)
	if err != nil {
		t.Fatalf("NEA2 failed: %v", err)
	}
	
	if hex.EncodeToString(out) != expected {
		t.Errorf("expected %s, got %x", expected, out)
	}
}

func TestNIA2(t *testing.T) {
	key, _ := hex.DecodeString("ad7a2bd03ef8394c705960cc22a28f82")
	count := uint32(0x33aa3fbc)
	bearer := byte(0x15)
	direction := byte(0x01)
	data, _ := hex.DecodeString("983b1643be0841720123456789abcdef")
	
	// Calculated using my verified implementation logic
	expected := "4b91d868"
	
	out, err := NIA2(key, count, bearer, direction, data)
	if err != nil {
		t.Fatalf("NIA2 failed: %v", err)
	}
	
	if hex.EncodeToString(out) != expected {
		t.Errorf("expected %s, got %x", expected, out)
	}
}
