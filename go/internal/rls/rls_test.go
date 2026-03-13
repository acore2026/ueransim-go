package rls

import (
	"bytes"
	"testing"
)

func TestRlsPdu(t *testing.T) {
	m := &RlsMessage{
		MsgType: PDU_TRANSMISSION,
		Sti:     0x1122334455667788,
		PduType: PDU_TYPE_RRC,
		PduId:   123,
		Payload: 456,
		Pdu:     []byte{0xDE, 0xAD, 0xBE, 0xEF},
	}

	encoded, err := m.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.MsgType != m.MsgType || decoded.Sti != m.Sti || decoded.PduType != m.PduType || decoded.PduId != m.PduId || !bytes.Equal(decoded.Pdu, m.Pdu) {
		t.Errorf("Decoded message mismatch: %+v", decoded)
	}
}
