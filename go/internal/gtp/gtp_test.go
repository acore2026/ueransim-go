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
