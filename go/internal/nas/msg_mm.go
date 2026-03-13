package nas

import (
	"github.com/acore2026/ueransim-go/internal/utils"
)

// RegistrationRequest message
type RegistrationRequest struct {
	// Mandatory
	RegistrationType     IE5gsRegistrationType
	NasKeySetIdentifier  IENasKeySetIdentifier
	MobileIdentity       IE5gsMobileIdentity
	
	// Optional (simplified for start)
	UeSecurityCapability *UeSecurityCapability
}

func (m *RegistrationRequest) Encode() *utils.Buffer {
	b := utils.NewEmptyBuffer()
	
	// Protocol Discriminator
	b.AppendByte(PD_5G_MOBILITY_MANAGEMENT)
	
	// Security Header Type (0: Plain NAS)
	b.AppendByte(0x00)
	
	// Message Type
	b.AppendByte(MsgTypeRegistrationRequest)
	
	// Type 1 IEs are combined (Registration Type + NAS Key Set Identifier)
	// Byte 1: RegType(4 bits) | NasKeySet(4 bits)
	octet := (m.NasKeySetIdentifier.Encode() << 4) | m.RegistrationType.Encode()
	b.AppendByte(octet)
	
	// Mobile Identity (TLV)
	m.MobileIdentity.Encode(b)
	
	// Optional IEs would go here with their IEI
	if m.UeSecurityCapability != nil {
		b.AppendByte(0x2E) // IEI for UESecurityCapability
		m.UeSecurityCapability.Encode(b)
	}
	
	return b
}

type UeSecurityCapability struct {
	EA0, EA1, EA2, EA3, EA4, EA5, EA6, EA7 bool
	IA0, IA1, IA2, IA3, IA4, IA5, IA6, IA7 bool
}

func (u *UeSecurityCapability) Encode(b *utils.Buffer) {
	// Length (fixed 2 bytes for now)
	b.AppendByte(0x02)
	
	var ea byte
	if u.EA0 { ea |= 0x80 }
	if u.EA1 { ea |= 0x40 }
	if u.EA2 { ea |= 0x20 }
	if u.EA3 { ea |= 0x10 }
	b.AppendByte(ea)
	
	var ia byte
	if u.IA0 { ia |= 0x80 }
	if u.IA1 { ia |= 0x40 }
	if u.IA2 { ia |= 0x20 }
	if u.IA3 { ia |= 0x10 }
	b.AppendByte(ia)
}
