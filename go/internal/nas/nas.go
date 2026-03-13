package nas

import (
	"github.com/acore2026/ueransim-go/internal/utils"
)

// ProtocolDiscriminator values
const (
	PD_5G_MOBILITY_MANAGEMENT byte = 0x7E
	PD_5G_SESSION_MANAGEMENT  byte = 0x2E
)

// SecurityHeaderType values
type SecurityHeaderType byte

const (
	SecurityHeaderTypePlainNas                                  SecurityHeaderType = 0
	SecurityHeaderTypeIntegrityProtected                        SecurityHeaderType = 1
	SecurityHeaderTypeIntegrityProtectedAndCiphered             SecurityHeaderType = 2
	SecurityHeaderTypeIntegrityProtectedWithNewSecurityContext  SecurityHeaderType = 3
	SecurityHeaderTypeIntegrityProtectedAndCipheredWithNewSecurityContext SecurityHeaderType = 4
)

// MessageType values (simplified for start)
const (
	MsgTypeRegistrationRequest      byte = 0x41
	MsgTypeRegistrationAccept       byte = 0x42
	MsgTypeRegistrationComplete     byte = 0x43
	MsgTypeAuthenticationRequest     byte = 0x56
	MsgTypeAuthenticationResponse    byte = 0x57
	MsgTypeSecurityModeCommand      byte = 0x5D
	MsgTypeSecurityModeComplete     byte = 0x5E
)

// IE is the interface for all Information Elements
type IE interface {
	Encode(b *utils.Buffer)
	Decode(b *utils.BitBuffer, length int) error
}

// IE1 (Half-octet) is a special case in 3GPP
type IE1 interface {
	Encode() byte // Returns the 4-bit value
}

// IEType defines the IE encoding format
type IEType int

const (
	IEType1 IEType = 1 // TV (Half-octet)
	IEType2 IEType = 2 // T (1 octet)
	IEType3 IEType = 3 // TV (Fixed length)
	IEType4 IEType = 4 // TLV (1-octet length)
	IEType6 IEType = 6 // TLV (2-octet length)
)
