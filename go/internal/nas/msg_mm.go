package nas

import (
	"fmt"
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

// AuthenticationRequest message
type AuthenticationRequest struct {
	NasKeySetIdentifier IENasKeySetIdentifier
	Rand                [16]byte
	Autn                [16]byte
}

func DecodeAuthenticationRequest(data []byte) (*AuthenticationRequest, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("NAS PDU too short")
	}
	if data[2] != MsgTypeAuthenticationRequest {
		return nil, fmt.Errorf("not an Authentication Request: 0x%02x", data[2])
	}
	
	m := &AuthenticationRequest{}
	m.NasKeySetIdentifier.KeySetIdentifier = data[3] & 0x07
	
	offset := 4
	// ABBA (Mandatory, LV)
	abbaLen := int(data[offset])
	offset += 1 + abbaLen
	
	// RAND (Mandatory, TV, IEI=0x21)
	if len(data) < offset+17 {
		return nil, fmt.Errorf("missing RAND at offset %d", offset)
	}
	if data[offset] != 0x21 {
		return nil, fmt.Errorf("invalid RAND IEI: 0x%02x", data[offset])
	}
	copy(m.Rand[:], data[offset+1:offset+17])
	offset += 17
	
	// AUTN (Mandatory, TLV, IEI=0x20)
	if len(data) < offset+18 {
		return nil, fmt.Errorf("missing AUTN at offset %d", offset)
	}
	if data[offset] != 0x20 {
		return nil, fmt.Errorf("invalid AUTN IEI: 0x%02x", data[offset])
	}
	autnLen := int(data[offset+1])
	if autnLen != 16 {
		return nil, fmt.Errorf("invalid AUTN length: %d", autnLen)
	}
	copy(m.Autn[:], data[offset+2:offset+18])
	
	return m, nil
}

// AuthenticationResponse message
type AuthenticationResponse struct {
	ResStar []byte
}

func (m *AuthenticationResponse) Encode() *utils.Buffer {
	b := utils.NewEmptyBuffer()
	b.AppendByte(PD_5G_MOBILITY_MANAGEMENT)
	b.AppendByte(0x00) // Plain NAS
	b.AppendByte(MsgTypeAuthenticationResponse)
	
	// Authentication response parameter (TLV, IEI=0x2D)
	if len(m.ResStar) > 0 {
		b.AppendByte(0x2D)
		b.AppendByte(byte(len(m.ResStar)))
		b.AppendBytes(m.ResStar)
	}
	
	return b
}
