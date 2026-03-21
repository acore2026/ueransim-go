package rlc

import (
	"testing"
)

func TestUMHeader(t *testing.T) {
	tests := []struct {
		si uint8
		sn uint8
		expected byte
	}{
		{0, 0, 0x00},
		{0, 63, 0x3F},
		{1, 0, 0x40},
		{2, 10, 0x8A},
		{3, 63, 0xFF},
	}

	for _, tt := range tests {
		h := UMHeader{SI: tt.si, SN: tt.sn}
		encoded := h.Encode()
		if encoded != tt.expected {
			t.Errorf("Encode(%d, %d) = 0x%02x, want 0x%02x", tt.si, tt.sn, encoded, tt.expected)
		}

		decoded := DecodeUMHeader(encoded)
		if decoded.SI != tt.si || decoded.SN != tt.sn {
			t.Errorf("Decode(0x%02x) = {%d, %d}, want {%d, %d}", encoded, decoded.SI, decoded.SN, tt.si, tt.sn)
		}
	}
}
