package rrc

import (
	"github.com/acore2026/ueransim-go/internal/utils"
)

// Minimal RRC Builder using manual bit packing for essential messages.
// This is a placeholder for a full UPER implementation.

func BuildRRCSetupRequest(ueIdentity uint64) []byte {
	bs := utils.NewBitString()
	
	// UL-CCCH-Message ::= SEQUENCE { message Choice }
	// choice index 0 (initiatingMessage)
	bs.WriteBits(0, 1) 
	
	// initiatingMessage choice index 0 (rrcSetupRequest)
	bs.WriteBits(0, 1) 
	
	// RRCSetupRequest ::= SEQUENCE { rrcSetupRequest-IEs RRCSetupRequest-IEs }
	// ue-Identity choice index 0 (randomValue)
	bs.WriteBits(0, 1)
	
	// randomValue BIT STRING (SIZE (39))
	bs.WriteBits(int(ueIdentity>>32), 7)
	bs.WriteBits(int(ueIdentity&0xFFFFFFFF), 32)
	
	// establishmentCause ENUMERATED (8 values)
	// mo-Signalling (index 3)
	bs.WriteBits(3, 3)
	
	// spare BIT STRING (SIZE (1))
	bs.WriteBits(0, 1)
	
	return bs.Data()
}

func BuildRRCSetupComplete(nasPdu []byte) []byte {
	bs := utils.NewBitString()

	// UL-DCCH-Message ::= SEQUENCE { message Choice }
	// message Choice Index 0: c1
	bs.WriteBits(0, 1)

	// c1 choice index 2 (rrcSetupComplete). 16 options -> 4 bits.
	bs.WriteBits(2, 4)

	// RRCSetupComplete ::= SEQUENCE { rrc-TransactionIdentifier, criticalExtensions Choice }
	// rrc-TransactionIdentifier = 0 (2 bits)
	bs.WriteBits(0, 2)

	// criticalExtensions choice index 0 (rrcSetupComplete)
	bs.WriteBits(0, 1)

	// RRCSetupComplete-IEs ::= SEQUENCE { ... }
	// selectedPLMN-Identity = 1 (index 0). INTEGER (1..12) -> 4 bits.
	bs.WriteBits(0, 4)

	// registeredAMF OPTIONAL (absent)
	bs.WriteBits(0, 1)

	// guami-Type OPTIONAL (absent)
	bs.WriteBits(0, 1)

	// s-NSSAI-List OPTIONAL (absent)
	bs.WriteBits(0, 1)

	// dedicatedNAS-Message (OCTET STRING)
	// UPER length determinant for small length (<128)
	bs.WriteBits(len(nasPdu), 8)
	for _, b := range nasPdu {
		bs.WriteBits(int(b), 8)
	}

	// ng-5G-S-TMSI-Value OPTIONAL (absent)
	bs.WriteBits(0, 1)

	// lateNonCriticalExtension OPTIONAL (absent)
	bs.WriteBits(0, 1)

	// nonCriticalExtension OPTIONAL (absent)
	bs.WriteBits(0, 1)

	return bs.Data()
}

func BuildULInformationTransfer(nasPdu []byte) []byte {
	bs := utils.NewBitString()

	// UL-DCCH-Message ::= SEQUENCE { message Choice }
	// message Choice Index 0: c1
	bs.WriteBits(0, 1)

	// c1 choice index 7 (ulInformationTransfer). 16 options -> 4 bits.
	bs.WriteBits(7, 4)

	// ULInformationTransfer ::= SEQUENCE { criticalExtensions Choice }
	// criticalExtensions choice index 0 (ulInformationTransfer)
	bs.WriteBits(0, 1)

	// ULInformationTransfer-IEs ::= SEQUENCE { ... }
	// dedicatedNAS-Message (OCTET STRING)
	bs.WriteBits(len(nasPdu), 8)
	for _, b := range nasPdu {
		bs.WriteBits(int(b), 8)
	}

	// lateNonCriticalExtension OPTIONAL (absent)
	bs.WriteBits(0, 1)

	// nonCriticalExtension OPTIONAL (absent)
	bs.WriteBits(0, 1)

	return bs.Data()
}

func BuildDLInformationTransfer(nasPdu []byte) []byte {
	bs := utils.NewBitString()

	// DL-DCCH-Message ::= SEQUENCE { message Choice }
	// message Choice Index 0: c1
	bs.WriteBits(0, 1)

	// c1 choice index 5 (dlInformationTransfer). 16 options -> 4 bits.
	bs.WriteBits(5, 4)

	// DLInformationTransfer ::= SEQUENCE { rrc-TransactionIdentifier, criticalExtensions Choice }
	// rrc-TransactionIdentifier = 0 (2 bits)
	bs.WriteBits(0, 2)

	// criticalExtensions choice index 0 (dlInformationTransfer)
	bs.WriteBits(0, 1)

	// DLInformationTransfer-IEs ::= SEQUENCE { ... }
	// dedicatedNAS-Message (OCTET STRING)
	bs.WriteBits(len(nasPdu), 8)
	for _, b := range nasPdu {
		bs.WriteBits(int(b), 8)
	}

	// lateNonCriticalExtension OPTIONAL (absent)
	bs.WriteBits(0, 1)

	// nonCriticalExtension OPTIONAL (absent)
	bs.WriteBits(0, 1)

	return bs.Data()
}
