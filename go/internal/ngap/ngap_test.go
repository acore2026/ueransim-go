package ngap

import (
	"encoding/hex"
	"testing"
)

func TestEncodeNGSetupRequest(t *testing.T) {
	gnbName := "ueransim-gnb"
	gnbID := []byte{0x00, 0x01, 0x02}
	plmnID := []byte{0x02, 0xf8, 0x39}

	pdu, err := BuildNGSetupRequest(gnbName, gnbID, 24, plmnID)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	res, err := Encode(pdu)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	t.Logf("Encoded hex: %s", hex.EncodeToString(res))
	
	if len(res) == 0 {
		t.Error("encoded result is empty")
	}
}

func TestInitialUEMessage(t *testing.T) {
	plmnID := []byte{0x02, 0xf8, 0x39}
	tac := []byte{0x00, 0x00, 0x01}
	nrCellID := []byte{0x00, 0x00, 0x00, 0x00, 0x10}
	nasPdu := []byte{0x7e, 0x00, 0x41, 0x79, 0x00, 0x0d, 0x01, 0x02, 0xf8, 0x39, 0x00, 0x00, 0x00, 0x00, 0x01, 0x23, 0x45, 0x67, 0x89}

	userLocation := BuildUserLocationInformationNR(plmnID, tac, nrCellID)
	pdu, err := BuildInitialUEMessage(1, nasPdu, userLocation)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	res, err := Encode(pdu)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	t.Logf("Encoded hex: %s", hex.EncodeToString(res))

	if len(res) == 0 {
		t.Error("encoded result is empty")
	}
}
