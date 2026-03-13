package utils

import (
	"bytes"
	"testing"
)

func TestBuffer(t *testing.T) {
	b := NewEmptyBuffer()
	b.AppendByte(0x11)
	b.AppendUint16(0x2233)
	b.AppendUint24(0x445566)
	b.AppendUint32(0x778899AA)

	expected := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA}
	if !bytes.Equal(b.Data(), expected) {
		t.Errorf("expected %X, got %X", expected, b.Data())
	}

	if b.GetUint16(1) != 0x2233 {
		t.Errorf("expected 0x2233, got %X", b.GetUint16(1))
	}
}
