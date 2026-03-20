package nas

import (
	"bytes"
	"fmt"

	extmsg "github.com/acore2026/nas/nasMessage"
	exttype "github.com/acore2026/nas/nasType"

	"github.com/acore2026/ueransim-go/internal/utils"
)

// The adapter boundary intentionally keeps procedure code using the local helper
// API while delegating repetitive message/IE structure handling to the forked NAS
// module.
//
// Current happy-path mapping:
// - RegistrationRequest                -> nasMessage.RegistrationRequest
// - AuthenticationRequest             -> nasMessage.AuthenticationRequest
// - AuthenticationResponse            -> nasMessage.AuthenticationResponse
// - IdentityRequest                   -> nasMessage.IdentityRequest
// - IdentityResponse                  -> nasMessage.IdentityResponse
// - SecurityModeCommand               -> nasMessage.SecurityModeCommand
// - SecurityModeComplete              -> nasMessage.SecurityModeComplete
// - RegistrationComplete              -> nasMessage.RegistrationComplete
// - UlNasTransport                    -> nasMessage.ULNASTransport
// - DlNasTransport                    -> nasMessage.DLNASTransport
// - PduSessionEstablishmentRequest    -> nasMessage.PDUSessionEstablishmentRequest
// - PduSessionEstablishmentAccept     -> nasMessage.PDUSessionEstablishmentAccept

func encodeWithBuilder(build func(*bytes.Buffer) error) *utils.Buffer {
	var raw bytes.Buffer
	if err := build(&raw); err != nil {
		panic(err)
	}
	b := utils.NewEmptyBuffer()
	b.AppendBytes(raw.Bytes())
	return b
}

func buildMobileIdentityPayload(identity IE5gsMobileIdentity) ([]byte, error) {
	b := utils.NewEmptyBuffer()
	identity.Encode(b)
	data := b.Data()
	if len(data) < 2 {
		return nil, fmt.Errorf("mobile identity payload too short")
	}
	length := int(data[0])<<8 | int(data[1])
	if len(data) != 2+length {
		return nil, fmt.Errorf("mobile identity length mismatch")
	}
	return append([]byte(nil), data[2:]...), nil
}

func buildMobileIdentity5GS(identity IE5gsMobileIdentity) (*exttype.MobileIdentity5GS, error) {
	payload, err := buildMobileIdentityPayload(identity)
	if err != nil {
		return nil, err
	}
	mi := exttype.NewMobileIdentity5GS(0)
	mi.SetLen(uint16(len(payload)))
	copy(mi.Buffer, payload)
	return mi, nil
}

func buildMobileIdentity(identity IE5gsMobileIdentity) (*exttype.MobileIdentity, error) {
	payload, err := buildMobileIdentityPayload(identity)
	if err != nil {
		return nil, err
	}
	mi := exttype.NewMobileIdentity(0)
	mi.SetLen(uint16(len(payload)))
	copy(mi.Buffer, payload)
	return mi, nil
}

func buildCapability5GMM(cap *Capability5GMM) *exttype.Capability5GMM {
	if cap == nil {
		return nil
	}
	out := exttype.NewCapability5GMM(extmsg.RegistrationRequestCapability5GMMType)
	out.SetLen(byte(len(cap.Octets)))
	copy(out.Octet[:], cap.Octets[:])
	return out
}

func buildUESecurityCapability(sec *UeSecurityCapability) *exttype.UESecurityCapability {
	if sec == nil {
		return nil
	}
	out := exttype.NewUESecurityCapability(extmsg.RegistrationRequestUESecurityCapabilityType)
	out.SetLen(2)
	if sec.EA0 {
		out.SetEA0_5G(1)
	}
	if sec.EA1 {
		out.SetEA1_128_5G(1)
	}
	if sec.EA2 {
		out.SetEA2_128_5G(1)
	}
	if sec.EA3 {
		out.SetEA3_128_5G(1)
	}
	if sec.EA4 {
		out.SetEA4_5G(1)
	}
	if sec.EA5 {
		out.SetEA5_5G(1)
	}
	if sec.EA6 {
		out.SetEA6_5G(1)
	}
	if sec.EA7 {
		out.SetEA7_5G(1)
	}
	if sec.IA0 {
		out.SetIA0_5G(1)
	}
	if sec.IA1 {
		out.SetIA1_128_5G(1)
	}
	if sec.IA2 {
		out.SetIA2_128_5G(1)
	}
	if sec.IA3 {
		out.SetIA3_128_5G(1)
	}
	if sec.IA4 {
		out.SetIA4_5G(1)
	}
	if sec.IA5 {
		out.SetIA5_5G(1)
	}
	if sec.IA6 {
		out.SetIA6_5G(1)
	}
	if sec.IA7 {
		out.SetIA7_5G(1)
	}
	return out
}

func buildSNSSAI(sn *SNssai, iei uint8) *exttype.SNSSAI {
	if sn == nil {
		return nil
	}
	out := exttype.NewSNSSAI(iei)
	if len(sn.SD) == 3 {
		out.SetLen(4)
		out.SetSST(sn.SST)
		var sd [3]uint8
		copy(sd[:], sn.SD)
		out.SetSD(sd)
		return out
	}
	out.SetLen(1)
	out.SetSST(sn.SST)
	return out
}
