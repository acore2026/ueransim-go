package nas

import (
	"fmt"
	"github.com/acore2026/ueransim-go/internal/utils"
)

// RegistrationRequest message
type RegistrationRequest struct {
	// Mandatory
	RegistrationType    IE5gsRegistrationType
	NasKeySetIdentifier IENasKeySetIdentifier
	MobileIdentity      IE5gsMobileIdentity

	// Optional (simplified for start)
	Capability5GMM       *Capability5GMM
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
	if m.Capability5GMM != nil {
		b.AppendByte(0x10)
		m.Capability5GMM.Encode(b)
	}
	if m.UeSecurityCapability != nil {
		b.AppendByte(0x2E) // IEI for UESecurityCapability
		m.UeSecurityCapability.Encode(b)
	}

	return b
}

type Capability5GMM struct {
	Octets [13]byte
}

func (c *Capability5GMM) Encode(b *utils.Buffer) {
	b.AppendByte(byte(len(c.Octets)))
	b.AppendBytes(c.Octets[:])
}

type UeSecurityCapability struct {
	EA0, EA1, EA2, EA3, EA4, EA5, EA6, EA7 bool
	IA0, IA1, IA2, IA3, IA4, IA5, IA6, IA7 bool
}

func (u *UeSecurityCapability) Encode(b *utils.Buffer) {
	// Length (fixed 2 bytes for now)
	b.AppendByte(0x02)

	var ea byte
	if u.EA0 {
		ea |= 0x80
	}
	if u.EA1 {
		ea |= 0x40
	}
	if u.EA2 {
		ea |= 0x20
	}
	if u.EA3 {
		ea |= 0x10
	}
	b.AppendByte(ea)

	var ia byte
	if u.IA0 {
		ia |= 0x80
	}
	if u.IA1 {
		ia |= 0x40
	}
	if u.IA2 {
		ia |= 0x20
	}
	if u.IA3 {
		ia |= 0x10
	}
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
	// Based on hex: 02 00 00
	abbaLen := int(data[offset])
	offset += 1 + abbaLen

	// RAND (Mandatory, TV, IEI=0x21)
	if len(data) < offset+17 {
		return nil, fmt.Errorf("missing RAND at offset %d, total len %d", offset, len(data))
	}
	if data[offset] != 0x21 {
		return nil, fmt.Errorf("invalid RAND IEI: 0x%02x at offset %d", data[offset], offset)
	}
	copy(m.Rand[:], data[offset+1:offset+17])
	offset += 17

	// AUTN (Mandatory, TLV, IEI=0x20)
	if len(data) < offset+18 {
		return nil, fmt.Errorf("missing AUTN at offset %d, total len %d", offset, len(data))
	}
	if data[offset] != 0x20 {
		return nil, fmt.Errorf("invalid AUTN IEI: 0x%02x at offset %d", data[offset], offset)
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

		resStar := m.ResStar
		if len(resStar) > 16 {
			resStar = resStar[len(resStar)-16:]
		}

		b.AppendByte(byte(len(resStar)))
		b.AppendBytes(resStar)
	}

	return b
}

type IdentityRequest struct {
	IdentityType byte
}

func DecodeIdentityRequest(data []byte) (*IdentityRequest, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("NAS PDU too short")
	}
	if data[2] != MsgTypeIdentityRequest {
		return nil, fmt.Errorf("not an Identity Request: 0x%02x", data[2])
	}
	return &IdentityRequest{
		IdentityType: data[3] & 0x07,
	}, nil
}

type IdentityResponse struct {
	MobileIdentity IE5gsMobileIdentity
}

func (m *IdentityResponse) Encode() *utils.Buffer {
	b := utils.NewEmptyBuffer()
	b.AppendByte(PD_5G_MOBILITY_MANAGEMENT)
	b.AppendByte(0x00)
	b.AppendByte(MsgTypeIdentityResponse)
	m.MobileIdentity.Encode(b)
	return b
}

// SecurityModeCommand message
type SecurityModeCommand struct {
	SelectedIntegrityAlgorithm byte
	SelectedCipheringAlgorithm byte
	NasKeySetIdentifier        byte
}

func DecodeSecurityModeCommand(data []byte) (*SecurityModeCommand, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("NAS PDU too short")
	}
	if data[2] != MsgTypeSecurityModeCommand {
		return nil, fmt.Errorf("not a Security Mode Command: 0x%02x", data[2])
	}

	m := &SecurityModeCommand{}
	// data[3]: Selected NAS security algorithms
	m.SelectedCipheringAlgorithm = (data[3] >> 4) & 0x07
	m.SelectedIntegrityAlgorithm = data[3] & 0x07

	// data[4]: bits 4-6 is NasKeySetIdentifier
	m.NasKeySetIdentifier = (data[4] >> 4) & 0x07

	return m, nil
}

// SecurityModeComplete message
type SecurityModeComplete struct {
	MobileIdentity IE5gsMobileIdentity
}

func (m *SecurityModeComplete) Encode() *utils.Buffer {
	b := utils.NewEmptyBuffer()
	b.AppendByte(PD_5G_MOBILITY_MANAGEMENT)
	b.AppendByte(0x00) // Will be updated by security layer
	b.AppendByte(MsgTypeSecurityModeComplete)

	// Mobile Identity (Mandatory, TLV)
	m.MobileIdentity.Encode(b)

	return b
}

// RegistrationComplete message
type RegistrationComplete struct{}

func (m *RegistrationComplete) Encode() *utils.Buffer {
	b := utils.NewEmptyBuffer()
	b.AppendByte(PD_5G_MOBILITY_MANAGEMENT)
	b.AppendByte(0x00)
	b.AppendByte(MsgTypeRegistrationComplete)
	return b
}

type UlNasTransport struct {
	PayloadContainerType byte
	PayloadContainer     []byte
	PduSessionID         byte
	RequestType          byte
	SNssai               *SNssai
	Dnn                  string
}

func (m *UlNasTransport) Encode() *utils.Buffer {
	b := utils.NewEmptyBuffer()
	b.AppendByte(PD_5G_MOBILITY_MANAGEMENT)
	b.AppendByte(0x00)
	b.AppendByte(MsgTypeUlNasTransport)
	b.AppendByte(m.PayloadContainerType & 0x0F)
	b.AppendUint16(uint16(len(m.PayloadContainer)))
	b.AppendBytes(m.PayloadContainer)

	if m.PduSessionID != 0 {
		b.AppendByte(0x12)
		b.AppendByte(0x01)
		b.AppendByte(m.PduSessionID)
	}

	if m.RequestType != 0 {
		b.AppendByte(0x80 | (m.RequestType & 0x0F))
	}

	if m.SNssai != nil {
		body := utils.NewEmptyBuffer()
		body.AppendByte(m.SNssai.SST)
		if len(m.SNssai.SD) == 3 {
			body.AppendBytes(m.SNssai.SD)
		}
		b.AppendByte(0x22)
		b.AppendByte(byte(body.Len()))
		b.Append(body)
	}

	if m.Dnn != "" {
		body := utils.NewEmptyBuffer()
		body.AppendByte(byte(len(m.Dnn)))
		body.AppendBytes([]byte(m.Dnn))
		b.AppendByte(0x25)
		b.AppendByte(byte(body.Len()))
		b.Append(body)
	}

	return b
}

type DlNasTransport struct {
	PayloadContainerType byte
	PayloadContainer     []byte
	PduSessionID         byte
}

func DecodeDlNasTransport(data []byte) (*DlNasTransport, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("DL NAS Transport too short")
	}
	if data[2] != MsgTypeDlNasTransport {
		return nil, fmt.Errorf("not a DL NAS Transport: 0x%02x", data[2])
	}

	m := &DlNasTransport{
		PayloadContainerType: data[3] & 0x0F,
	}
	containerLen := int(data[4])<<8 | int(data[5])
	if len(data) < 6+containerLen {
		return nil, fmt.Errorf("DL NAS Transport container truncated")
	}
	m.PayloadContainer = append([]byte(nil), data[6:6+containerLen]...)
	offset := 6 + containerLen

	for offset < len(data) {
		iei := data[offset]
		offset++
		switch iei {
		case 0x12:
			m.PduSessionID = data[offset]
			offset++
		default:
			return m, nil
		}
	}
	return m, nil
}

type SNssai struct {
	SST byte
	SD  []byte
}

type PduSessionEstablishmentRequest struct {
	PduSessionID   byte
	Pti            byte
	PduSessionType byte
	SscMode        byte
}

func (m *PduSessionEstablishmentRequest) Encode() *utils.Buffer {
	b := utils.NewEmptyBuffer()
	b.AppendByte(PD_5G_SESSION_MANAGEMENT)
	b.AppendByte(m.PduSessionID)
	b.AppendByte(m.Pti)
	b.AppendByte(MsgTypePduSessionEstablishmentRequest)
	b.AppendByte(0xFF)
	b.AppendByte(0xFF)
	b.AppendByte(0x90 | (m.PduSessionType & 0x0F))
	b.AppendByte(0xA0 | (m.SscMode & 0x0F))
	return b
}

type PduSessionEstablishmentAccept struct {
	PduSessionID byte
	Pti          byte
}

func DecodePduSessionEstablishmentAccept(data []byte) (*PduSessionEstablishmentAccept, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("PDU Session Establishment Accept too short")
	}
	if data[0] != PD_5G_SESSION_MANAGEMENT {
		return nil, fmt.Errorf("unexpected SM protocol discriminator 0x%02x", data[0])
	}
	if data[3] != MsgTypePduSessionEstablishmentAccept {
		return nil, fmt.Errorf("not a PDU Session Establishment Accept: 0x%02x", data[3])
	}
	return &PduSessionEstablishmentAccept{
		PduSessionID: data[1],
		Pti:          data[2],
	}, nil
}
