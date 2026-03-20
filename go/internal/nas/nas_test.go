package nas

import (
	"testing"
)

func TestRegistrationRequest(t *testing.T) {
	req := &RegistrationRequest{
		RegistrationType: IE5gsRegistrationType{
			FollowOnRequest:  true,
			RegistrationType: 0x01, // Initial registration
		},
		NasKeySetIdentifier: IENasKeySetIdentifier{
			KeySetIdentifier: 0x07, // No key
			Tsc:              false,
		},
		MobileIdentity: IE5gsMobileIdentity{
			Type: MobileIdentityTypeSuci,
			Suci: &Suci{
				Mcc:                    "208",
				Mnc:                    "93",
				Routing:                0x0000,
				Prot:                   0x00,
				HomeNetworkPublicKeyID: 0x00,
				MSIN:                   "0123456789",
			},
		},
	}

	buf := req.Encode()
	hex := buf.Hex()

	// Basic validation:
	// 7E (PD)
	// 00 (Security Header)
	// 41 (Message Type: Registration Request)
	// 79 (NAS Key Set ID 7 | Reg Type 9 (FollowOn + Initial))
	// 000D (Length of Mobile Identity: 13 bytes)
	// 01 (Type SUCI)
	// 02 F8 39 (MCC 208, MNC 93)
	// 00 00 (Routing)
	// 00 (Prot)
	// 00 (PubKeyID)
	// 01 23 45 67 89 (MSIN)

	expectedStart := "7e004179000d0102f839000000001032547698"
	if len(hex) < len(expectedStart) {
		t.Fatalf("hex too short: %s", hex)
	}
	if hex[:len(expectedStart)] != expectedStart {
		t.Errorf("expected start %s, got %s", expectedStart, hex)
	}
}

func TestRegistrationComplete(t *testing.T) {
	msg := (&RegistrationComplete{}).Encode().Hex()
	if msg != "7e0043" {
		t.Fatalf("expected registration complete hex 7e0043, got %s", msg)
	}
}

func TestDecodeIdentityRequest(t *testing.T) {
	msg, err := DecodeIdentityRequest([]byte{0x7e, 0x00, 0x5b, 0x05})
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if msg.IdentityType != MobileIdentityTypeImeisv {
		t.Fatalf("expected IMEISV request, got %d", msg.IdentityType)
	}
}

func TestIdentityResponseIMEISV(t *testing.T) {
	msg := (&IdentityResponse{
		MobileIdentity: IE5gsMobileIdentity{
			Type:   MobileIdentityTypeImeisv,
			Digits: "4370816125816151",
		},
	}).Encode().Hex()

	expected := "7e005c00094573806121856151f1"
	if msg != expected {
		t.Fatalf("expected %s, got %s", expected, msg)
	}
}

func TestUlNasTransportEncode(t *testing.T) {
	msg := (&UlNasTransport{
		PayloadContainerType: 1,
		PayloadContainer:     []byte{0x2e, 0x01, 0x01, 0xc1},
		PduSessionID:         1,
		RequestType:          1,
		SNssai: &SNssai{
			SST: 0x01,
			SD:  []byte{0x01, 0x02, 0x03},
		},
		Dnn: "internet",
	}).Encode().Hex()

	expected := "7e00670100042e0101c112010181220401010203250908696e7465726e6574"
	if msg != expected {
		t.Fatalf("expected %s, got %s", expected, msg)
	}
}

func TestDecodeDlNasTransport(t *testing.T) {
	data := []byte{
		0x7e, 0x00, 0x68, 0x01, 0x00, 0x04,
		0x2e, 0x01, 0x01, 0xc2,
		0x12, 0x01,
	}

	msg, err := DecodeDlNasTransport(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if msg.PayloadContainerType != 1 {
		t.Fatalf("expected payload container type 1, got %d", msg.PayloadContainerType)
	}
	if msg.PduSessionID != 1 {
		t.Fatalf("expected PDU session ID 1, got %d", msg.PduSessionID)
	}
}

func TestPduSessionEstablishmentAcceptDecode(t *testing.T) {
	msg, err := DecodePduSessionEstablishmentAccept([]byte{0x2e, 0x01, 0x01, 0xc2})
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if msg.PduSessionID != 1 || msg.Pti != 1 {
		t.Fatalf("unexpected accept payload: %+v", msg)
	}
}
