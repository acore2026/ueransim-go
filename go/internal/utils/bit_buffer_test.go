package utils

import (
	"bytes"
	"testing"
)

func TestBitBuffer(t *testing.T) {
	data := make([]byte, 2)
	bb := NewBitBuffer(data)

	// Write 10101010 11110000
	bb.WriteBits(0xAA, 8)
	bb.WriteBits(0xF0, 8)

	if !bytes.Equal(bb.Data(), []byte{0xAA, 0xF0}) {
		t.Errorf("expected [0xAA 0xF0], got %v", bb.Data())
	}

	bb.Seek(0)
	if v := bb.ReadBits(8); v != 0xAA {
		t.Errorf("expected 0xAA, got 0x%X", v)
	}
	if v := bb.ReadBits(4); v != 0xF {
		t.Errorf("expected 0xF, got 0x%X", v)
	}
	if v := bb.ReadBits(4); v != 0 {
		t.Errorf("expected 0, got 0x%X", v)
	}
}

func TestBitString(t *testing.T) {
	bs := NewBitString()
	bs.WriteBits(0x07, 3) // 111
	bs.WriteBits(0x00, 1) // 0
	
	if bs.BitLength() != 4 {
		t.Errorf("expected length 4, got %d", bs.BitLength())
	}
	
	// Final byte should be 1110 0000 = 0xE0
	if bs.Data()[0] != 0xE0 {
		t.Errorf("expected 0xE0, got 0x%X", bs.Data()[0])
	}
}
