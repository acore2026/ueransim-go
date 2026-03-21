package nas

import (
	"fmt"
	extnas "github.com/acore2026/nas"
	extmsg "github.com/acore2026/nas/nasMessage"
	exttype "github.com/acore2026/nas/nasType"
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
	identity, err := buildMobileIdentity5GS(m.MobileIdentity)
	if err != nil {
		panic(err)
	}
	msg := extmsg.NewRegistrationRequest(extnas.MsgTypeRegistrationRequest)
	msg.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(PD_5G_MOBILITY_MANAGEMENT)
	msg.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(extnas.SecurityHeaderTypePlainNas)
	msg.RegistrationRequestMessageIdentity.SetMessageType(extnas.MsgTypeRegistrationRequest)
	msg.NgksiAndRegistrationType5GS.SetFOR(boolToUint8(m.RegistrationType.FollowOnRequest))
	msg.NgksiAndRegistrationType5GS.SetRegistrationType5GS(m.RegistrationType.RegistrationType)
	msg.NgksiAndRegistrationType5GS.SetNasKeySetIdentifiler(m.NasKeySetIdentifier.KeySetIdentifier)
	msg.NgksiAndRegistrationType5GS.SetTSC(boolToUint8(m.NasKeySetIdentifier.Tsc))
	msg.MobileIdentity5GS = *identity
	msg.Capability5GMM = buildCapability5GMM(m.Capability5GMM)
	msg.UESecurityCapability = buildUESecurityCapability(m.UeSecurityCapability)
	return encodeWithBuilder(msg.EncodeRegistrationRequest)
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
	wire := append([]byte(nil), data...)
	msg := extnas.NewMessage()
	if err := msg.PlainNasDecode(&wire); err != nil {
		return nil, err
	}
	if msg.GmmMessage == nil || msg.GmmMessage.AuthenticationRequest == nil {
		return nil, fmt.Errorf("not an Authentication Request")
	}
	src := msg.GmmMessage.AuthenticationRequest
	if src.AuthenticationParameterRAND == nil || src.AuthenticationParameterAUTN == nil {
		return nil, fmt.Errorf("authentication request missing RAND/AUTN")
	}
	m := &AuthenticationRequest{}
	m.NasKeySetIdentifier.KeySetIdentifier = src.SpareHalfOctetAndNgksi.GetNasKeySetIdentifiler()
	copy(m.Rand[:], src.AuthenticationParameterRAND.Octet[:])
	copy(m.Autn[:], src.AuthenticationParameterAUTN.Octet[:])
	return m, nil
}

// AuthenticationResponse message
type AuthenticationResponse struct {
	ResStar []byte
}

func (m *AuthenticationResponse) Encode() *utils.Buffer {
	msg := extmsg.NewAuthenticationResponse(extnas.MsgTypeAuthenticationResponse)
	msg.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(PD_5G_MOBILITY_MANAGEMENT)
	msg.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(extnas.SecurityHeaderTypePlainNas)
	msg.AuthenticationResponseMessageIdentity.SetMessageType(extnas.MsgTypeAuthenticationResponse)
	if len(m.ResStar) > 0 {
		param := exttype.NewAuthenticationResponseParameter(extmsg.AuthenticationResponseAuthenticationResponseParameterType)
		param.SetLen(16)
		resStar := m.ResStar
		if len(resStar) > 16 {
			resStar = resStar[len(resStar)-16:]
		}
		var arr [16]byte
		copy(arr[16-len(resStar):], resStar)
		param.SetRES(arr)
		msg.AuthenticationResponseParameter = param
	}
	return encodeWithBuilder(msg.EncodeAuthenticationResponse)
}

type IdentityRequest struct {
	IdentityType byte
}

func DecodeIdentityRequest(data []byte) (*IdentityRequest, error) {
	wire := append([]byte(nil), data...)
	msg := extnas.NewMessage()
	if err := msg.PlainNasDecode(&wire); err != nil {
		return nil, err
	}
	if msg.GmmMessage == nil || msg.GmmMessage.IdentityRequest == nil {
		return nil, fmt.Errorf("not an Identity Request")
	}
	return &IdentityRequest{
		IdentityType: msg.GmmMessage.IdentityRequest.SpareHalfOctetAndIdentityType.GetTypeOfIdentity(),
	}, nil
}

type IdentityResponse struct {
	MobileIdentity IE5gsMobileIdentity
}

func (m *IdentityResponse) Encode() *utils.Buffer {
	identity, err := buildMobileIdentity(m.MobileIdentity)
	if err != nil {
		panic(err)
	}
	msg := extmsg.NewIdentityResponse(extnas.MsgTypeIdentityResponse)
	msg.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(PD_5G_MOBILITY_MANAGEMENT)
	msg.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(extnas.SecurityHeaderTypePlainNas)
	msg.IdentityResponseMessageIdentity.SetMessageType(extnas.MsgTypeIdentityResponse)
	msg.MobileIdentity = *identity
	return encodeWithBuilder(msg.EncodeIdentityResponse)
}

// SecurityModeCommand message
type SecurityModeCommand struct {
	SelectedIntegrityAlgorithm byte
	SelectedCipheringAlgorithm byte
	NasKeySetIdentifier        byte
}

func DecodeSecurityModeCommand(data []byte) (*SecurityModeCommand, error) {
	wire := append([]byte(nil), data...)
	msg := extnas.NewMessage()
	if err := msg.PlainNasDecode(&wire); err != nil {
		return nil, err
	}
	if msg.GmmMessage == nil || msg.GmmMessage.SecurityModeCommand == nil {
		return nil, fmt.Errorf("not a Security Mode Command")
	}
	src := msg.GmmMessage.SecurityModeCommand
	return &SecurityModeCommand{
		SelectedIntegrityAlgorithm: src.SelectedNASSecurityAlgorithms.GetTypeOfIntegrityProtectionAlgorithm(),
		SelectedCipheringAlgorithm: src.SelectedNASSecurityAlgorithms.GetTypeOfCipheringAlgorithm(),
		NasKeySetIdentifier:        src.SpareHalfOctetAndNgksi.GetNasKeySetIdentifiler(),
	}, nil
}

// SecurityModeComplete message
type SecurityModeComplete struct {
	MobileIdentity IE5gsMobileIdentity
}

func (m *SecurityModeComplete) Encode() *utils.Buffer {
	msg := extmsg.NewSecurityModeComplete(extnas.MsgTypeSecurityModeComplete)
	msg.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(PD_5G_MOBILITY_MANAGEMENT)
	msg.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(extnas.SecurityHeaderTypePlainNas)
	msg.SecurityModeCompleteMessageIdentity.SetMessageType(extnas.MsgTypeSecurityModeComplete)
	if m.MobileIdentity.Type == MobileIdentityTypeImeisv {
		payload, err := buildMobileIdentityPayload(m.MobileIdentity)
		if err != nil {
			panic(err)
		}
		imeisv := exttype.NewIMEISV(extmsg.SecurityModeCompleteIMEISVType)
		imeisv.SetLen(uint16(len(payload)))
		copy(imeisv.Octet[:], payload)
		msg.IMEISV = imeisv
	}
	return encodeWithBuilder(msg.EncodeSecurityModeComplete)
}

// RegistrationComplete message
type RegistrationComplete struct{}

func (m *RegistrationComplete) Encode() *utils.Buffer {
	msg := extmsg.NewRegistrationComplete(extnas.MsgTypeRegistrationComplete)
	msg.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(PD_5G_MOBILITY_MANAGEMENT)
	msg.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(extnas.SecurityHeaderTypePlainNas)
	msg.RegistrationCompleteMessageIdentity.SetMessageType(extnas.MsgTypeRegistrationComplete)
	return encodeWithBuilder(msg.EncodeRegistrationComplete)
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
	msg := extmsg.NewULNASTransport(extnas.MsgTypeULNASTransport)
	msg.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(PD_5G_MOBILITY_MANAGEMENT)
	msg.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(extnas.SecurityHeaderTypePlainNas)
	msg.ULNASTRANSPORTMessageIdentity.SetMessageType(extnas.MsgTypeULNASTransport)
	msg.SpareHalfOctetAndPayloadContainerType.SetPayloadContainerType(m.PayloadContainerType)
	msg.PayloadContainer.SetLen(uint16(len(m.PayloadContainer)))
	copy(msg.PayloadContainer.Buffer, m.PayloadContainer)
	if m.PduSessionID != 0 {
		ps := exttype.NewPduSessionID2Value(extmsg.ULNASTransportPduSessionID2ValueType)
		ps.SetPduSessionID2Value(m.PduSessionID)
		msg.PduSessionID2Value = ps
	}
	if m.RequestType != 0 {
		req := exttype.NewRequestType(extmsg.ULNASTransportRequestTypeType)
		req.SetRequestTypeValue(m.RequestType)
		msg.RequestType = req
	}
	msg.SNSSAI = buildSNSSAI(m.SNssai, extmsg.ULNASTransportSNSSAIType)
	if m.Dnn != "" {
		dnn := exttype.NewDNN(extmsg.ULNASTransportDNNType)
		dnn.SetDNN(m.Dnn)
		msg.DNN = dnn
	}
	return encodeWithBuilder(msg.EncodeULNASTransport)
}

type DlNasTransport struct {
	PayloadContainerType byte
	PayloadContainer     []byte
	PduSessionID         byte
}

func DecodeDlNasTransport(data []byte) (*DlNasTransport, error) {
	wire := append([]byte(nil), data...)
	msg := extnas.NewMessage()
	if err := msg.PlainNasDecode(&wire); err != nil {
		return nil, err
	}
	if msg.GmmMessage == nil || msg.GmmMessage.DLNASTransport == nil {
		return nil, fmt.Errorf("not a DL NAS Transport")
	}
	src := msg.GmmMessage.DLNASTransport
	out := &DlNasTransport{
		PayloadContainerType: src.SpareHalfOctetAndPayloadContainerType.GetPayloadContainerType(),
		PayloadContainer:     append([]byte(nil), src.PayloadContainer.Buffer...),
	}
	if src.PduSessionID2Value != nil {
		out.PduSessionID = src.PduSessionID2Value.GetPduSessionID2Value()
	}
	return out, nil
}

type SNssai struct {
	SST byte
	SD  []byte
}

func boolToUint8(v bool) uint8 {
	if v {
		return 1
	}
	return 0
}
