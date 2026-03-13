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
	
	expectedStart := "7e004179000d0102f839000000000123456789"
	if len(hex) < len(expectedStart) {
		t.Fatalf("hex too short: %s", hex)
	}
	if hex[:len(expectedStart)] != expectedStart {
		t.Errorf("expected start %s, got %s", expectedStart, hex)
	}
}
