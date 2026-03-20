package nas

import (
	"github.com/acore2026/ueransim-go/internal/utils"
)

// IE5gsRegistrationType (Type 1)
type IE5gsRegistrationType struct {
	FollowOnRequest  bool
	RegistrationType byte // 1: Initial, 2: Mobility, 3: Periodic, 4: Emergency
}

func (ie IE5gsRegistrationType) Encode() byte {
	val := ie.RegistrationType & 0x07
	if ie.FollowOnRequest {
		val |= 0x08
	}
	return val
}

// IENasKeySetIdentifier (Type 1)
type IENasKeySetIdentifier struct {
	KeySetIdentifier byte // 0..6, 7: No key
	Tsc              bool // True: mapped, False: native
}

func (ie IENasKeySetIdentifier) Encode() byte {
	val := ie.KeySetIdentifier & 0x07
	if ie.Tsc {
		val |= 0x08
	}
	return val
}

// IE5gsMobileIdentity (Type 4)
const (
	MobileIdentityTypeSuci   byte = 0x01
	MobileIdentityTypeGuti   byte = 0x02
	MobileIdentityTypeImei   byte = 0x03
	MobileIdentityTypeImeisv byte = 0x05
)

type IE5gsMobileIdentity struct {
	Type   byte
	Digits string
	Suci   *Suci
	Guti   *Guti
}

func (ie IE5gsMobileIdentity) Encode(b *utils.Buffer) {
	// Length is 1 octet after the Type octet if it's TLV
	// But MobileIdentity is often used as a mandatory IE without IEI.
	// We'll assume the caller handles the IEI if optional.

	temp := utils.NewEmptyBuffer()
	switch ie.Type {
	case MobileIdentityTypeSuci:
		if ie.Suci != nil {
			ie.Suci.Encode(temp)
		}
	case MobileIdentityTypeImei, MobileIdentityTypeImeisv:
		encodeDigitIdentity(temp, ie.Type, ie.Digits)
	case MobileIdentityTypeGuti:
		// TODO: Implement GUTI
	}

	b.AppendUint16(uint16(temp.Len()))
	b.Append(temp)
}

// Suci implementation
type Suci struct {
	Mcc                    string
	Mnc                    string
	Routing                uint16
	Prot                   byte // Protection Scheme: 0: Null, 1: Profile A, 2: Profile B
	HomeNetworkPublicKeyID byte
	MSIN                   string
}

func (s *Suci) Encode(b *utils.Buffer) {
	// 5GS Mobile Identity Type (SUCI)
	// Byte 1: Type (bit 1-3)
	b.AppendByte(MobileIdentityTypeSuci)

	// PLMN (MCC/MNC) - 3 bytes
	plmn := utils.EncodePlmn(s.Mcc, s.Mnc)
	b.AppendBytes(plmn)

	// Routing Indicator - 2 bytes
	b.AppendUint16(s.Routing)

	// Protection Scheme - 4 bits
	// Home Network Public Key ID - 1 byte
	b.AppendByte(s.Prot)
	b.AppendByte(s.HomeNetworkPublicKeyID)

	// MSIN (BCD encoded)
	msin := utils.EncodeBcd(s.MSIN)
	b.AppendBytes(msin)
}

type Guti struct {
	// TODO
}

func encodeDigitIdentity(b *utils.Buffer, identityType byte, digits string) {
	if len(digits) == 0 {
		b.AppendByte(identityType & 0x07)
		return
	}

	firstOctet := ((digits[0] - '0') << 4) | (identityType & 0x07)
	if len(digits)%2 != 0 {
		firstOctet |= 0x08
	}
	b.AppendByte(firstOctet)
	if len(digits) > 1 {
		b.AppendBytes(utils.EncodeBcd(digits[1:]))
	}
}
