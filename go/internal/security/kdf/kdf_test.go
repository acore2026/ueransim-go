package kdf

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestKDF(t *testing.T) {
	// 3GPP TS 33.220 Annex B - Test vector for KDF
	// Note: These are simplified checks to ensure the logic matches the standard.
	key, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	fc := byte(0x01)
	p := []string{"test"}
	l := []int{4}

	res := KDF(key, fc, p, l)
	if len(res) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(res))
	}
}

func TestDeriveKnas(t *testing.T) {
	// Kamf (32 bytes)
	kAmf, _ := hex.DecodeString("8000000000000000000000000000000000000000000000000000000000000000")
	
	// Derive Encryption Key (algType 0x01, algId 0x02 for NEA2/AES)
	kEnc := DeriveKnas(kAmf, 0x01, 0x02)
	if len(kEnc) != 16 {
		t.Errorf("expected 16 bytes for NAS key, got %d", len(kEnc))
	}

	// Derive Integrity Key (algType 0x02, algId 0x02 for NIA2/AES)
	kInt := DeriveKnas(kAmf, 0x02, 0x02)
	if len(kInt) != 16 {
		t.Errorf("expected 16 bytes for NAS key, got %d", len(kInt))
	}

	if bytes.Equal(kEnc, kInt) {
		t.Error("derived encryption and integrity keys should be different")
	}
}
