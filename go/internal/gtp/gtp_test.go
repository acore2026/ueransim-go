package gtp

import (
	"bytes"
	"testing"
)

func TestGtpPdu(t *testing.T) {
	m := &GtpMessage{
		SFlag:   true,
		MsgType: MT_G_PDU,
		Teid:    0x12345678,
		Seq:     100,
		Payload: []byte{0x11, 0x22, 0x33, 0x44},
	}

	encoded, err := m.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.MsgType != m.MsgType || decoded.Teid != m.Teid || decoded.Seq != m.Seq || !bytes.Equal(decoded.Payload, m.Payload) {
		t.Errorf("Decoded message mismatch: %+v", decoded)
	}
}

func TestGtpPduWithPduSessionContainer(t *testing.T) {
	qfi := uint8(9)
	m := &GtpMessage{
		EFlag:   true,
		MsgType: MT_G_PDU,
		Teid:    0x01020304,
		QFI:     &qfi,
		Payload: []byte{0x45, 0x00, 0x00, 0x14},
	}

	encoded, err := m.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if !decoded.EFlag {
		t.Fatalf("expected E flag to be set")
	}
	if decoded.QFI == nil || *decoded.QFI != qfi {
		t.Fatalf("unexpected QFI: %+v", decoded.QFI)
	}
	if decoded.MsgType != m.MsgType || decoded.Teid != m.Teid || !bytes.Equal(decoded.Payload, m.Payload) {
		t.Errorf("Decoded message mismatch: %+v", decoded)
	}
}
