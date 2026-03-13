package nas

import (
	"github.com/acore2026/ueransim-go/internal/utils"
)

// IE5gsRegistrationType (Type 1)
type IE5gsRegistrationType struct {
	FollowOnRequest      bool
	RegistrationType     byte // 1: Initial, 2: Mobility, 3: Periodic, 4: Emergency
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
	MobileIdentityTypeSuci byte = 0x01
	MobileIdentityTypeGuti byte = 0x02
	MobileIdentityTypeImei byte = 0x03
)

type IE5gsMobileIdentity struct {
	Type   byte
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
	case MobileIdentityTypeGuti:
		// TODO: Implement GUTI
	}
	
	b.AppendUint16(uint16(temp.Len()))
	b.Append(temp)
}

// Suci implementation
type Suci struct {
	Mcc      string
	Mnc      string
	Routing  uint16
	Prot     byte // Protection Scheme: 0: Null, 1: Profile A, 2: Profile B
	HomeNetworkPublicKeyID byte
	MSIN     string
}

func (s *Suci) Encode(b *utils.Buffer) {
	// 5GS Mobile Identity Type (SUCI)
	// Byte 1: Type (bit 1-3)
	b.AppendByte(MobileIdentityTypeSuci)
	
	// PLMN (MCC/MNC) - 3 bytes
	plmn := encodePlmn(s.Mcc, s.Mnc)
	b.AppendBytes(plmn)
	
	// Routing Indicator - 2 bytes
	b.AppendUint16(s.Routing)
	
	// Protection Scheme - 4 bits
	// Home Network Public Key ID - 1 byte
	b.AppendByte(s.Prot)
	b.AppendByte(s.HomeNetworkPublicKeyID)
	
	// MSIN (BCD encoded)
	msin := encodeBcd(s.MSIN)
	b.AppendBytes(msin)
}

func encodePlmn(mcc, mnc string) []byte {
	res := make([]byte, 3)
	// MCC 1, 2, 3
	// MNC 1, 2, (3)
	// Octet 1: MCC2 | MCC1
	// Octet 2: MNC3 | MCC3 (if MNC length is 3) or 0xF | MCC3
	// Octet 3: MNC2 | MNC1
	
	res[0] = (mcc[1]-'0')<<4 | (mcc[0]-'0')
	if len(mnc) == 3 {
		res[1] = (mnc[2]-'0')<<4 | (mcc[2]-'0')
	} else {
		res[1] = 0xF0 | (mcc[2]-'0')
	}
	res[2] = (mnc[1]-'0')<<4 | (mnc[0]-'0')
	return res
}

func encodeBcd(s string) []byte {
	res := make([]byte, (len(s)+1)/2)
	for i := 0; i < len(s); i++ {
		val := s[i] - '0'
		if i%2 == 0 {
			res[i/2] |= (val << 4)
		} else {
			res[i/2] |= val
		}
	}
	// If odd length, fill last nibble with 0xF
	if len(s)%2 != 0 {
		res[len(res)-1] |= 0x0F
	}
	return res
}

type Guti struct {
	// TODO
}
