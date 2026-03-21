package nas

import (
	"fmt"
	extnas "github.com/acore2026/nas"
	extmsg "github.com/acore2026/nas/nasMessage"
	exttype "github.com/acore2026/nas/nasType"
	"github.com/acore2026/ueransim-go/internal/utils"
	"net"
)

type PduSessionEstablishmentRequest struct {
	PduSessionID   byte
	Pti            byte
	PduSessionType byte
	SscMode        byte
}

func (m *PduSessionEstablishmentRequest) Encode() *utils.Buffer {
	msg := extmsg.NewPDUSessionEstablishmentRequest(extnas.MsgTypePDUSessionEstablishmentRequest)
	msg.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(PD_5G_SESSION_MANAGEMENT)
	msg.PDUSessionID.SetPDUSessionID(m.PduSessionID)
	msg.PTI.SetPTI(m.Pti)
	msg.PDUSESSIONESTABLISHMENTREQUESTMessageIdentity.SetMessageType(extnas.MsgTypePDUSessionEstablishmentRequest)
	msg.IntegrityProtectionMaximumDataRate.Octet = [2]uint8{0xff, 0xff}
	pduType := exttype.NewPDUSessionType(extmsg.PDUSessionEstablishmentRequestPDUSessionTypeType)
	pduType.SetPDUSessionTypeValue(m.PduSessionType)
	msg.PDUSessionType = pduType
	ssc := exttype.NewSSCMode(extmsg.PDUSessionEstablishmentRequestSSCModeType)
	ssc.SetSSCMode(m.SscMode)
	msg.SSCMode = ssc
	return encodeWithBuilder(msg.EncodePDUSessionEstablishmentRequest)
}

type PduSessionEstablishmentAccept struct {
	PduSessionID byte
	Pti          byte
	PDUAddress   string
}

func DecodePduSessionEstablishmentAccept(data []byte) (*PduSessionEstablishmentAccept, error) {
	wire := append([]byte(nil), data...)
	msg := extnas.NewMessage()
	if err := msg.PlainNasDecode(&wire); err != nil {
		return nil, err
	}
	if msg.GsmMessage == nil || msg.GsmMessage.PDUSessionEstablishmentAccept == nil {
		return nil, fmt.Errorf("not a PDU Session Establishment Accept")
	}
	src := msg.GsmMessage.PDUSessionEstablishmentAccept
	res := &PduSessionEstablishmentAccept{
		PduSessionID: src.PDUSessionID.GetPDUSessionID(),
		Pti:          src.PTI.GetPTI(),
	}
	if src.PDUAddress != nil && src.PDUAddress.GetLen() >= 5 && src.PDUAddress.GetPDUSessionTypeValue() == 1 {
		info := src.PDUAddress.GetPDUAddressInformation()
		res.PDUAddress = net.IPv4(info[0], info[1], info[2], info[3]).String()
	}
	return res, nil
}

type PduSessionModificationRequest struct {
	PduSessionID byte
	Pti          byte
}

func (m *PduSessionModificationRequest) Encode() *utils.Buffer {
	msg := extmsg.NewPDUSessionModificationRequest(extnas.MsgTypePDUSessionModificationRequest)
	msg.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(PD_5G_SESSION_MANAGEMENT)
	msg.PDUSessionID.SetPDUSessionID(m.PduSessionID)
	msg.PTI.SetPTI(m.Pti)
	msg.PDUSESSIONMODIFICATIONREQUESTMessageIdentity.SetMessageType(extnas.MsgTypePDUSessionModificationRequest)
	return encodeWithBuilder(msg.EncodePDUSessionModificationRequest)
}

type PduSessionModificationCommand struct {
	PduSessionID byte
	Pti          byte
}

func DecodePduSessionModificationCommand(data []byte) (*PduSessionModificationCommand, error) {
	wire := append([]byte(nil), data...)
	msg := extnas.NewMessage()
	if err := msg.PlainNasDecode(&wire); err != nil {
		return nil, err
	}
	if msg.GsmMessage == nil || msg.GsmMessage.PDUSessionModificationCommand == nil {
		return nil, fmt.Errorf("not a PDU Session Modification Command")
	}
	src := msg.GsmMessage.PDUSessionModificationCommand
	return &PduSessionModificationCommand{
		PduSessionID: src.PDUSessionID.GetPDUSessionID(),
		Pti:          src.PTI.GetPTI(),
	}, nil
}

type PduSessionModificationComplete struct {
	PduSessionID byte
	Pti          byte
}

func (m *PduSessionModificationComplete) Encode() *utils.Buffer {
	msg := extmsg.NewPDUSessionModificationComplete(extnas.MsgTypePDUSessionModificationComplete)
	msg.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(PD_5G_SESSION_MANAGEMENT)
	msg.PDUSessionID.SetPDUSessionID(m.PduSessionID)
	msg.PTI.SetPTI(m.Pti)
	msg.PDUSESSIONMODIFICATIONCOMPLETEMessageIdentity.SetMessageType(extnas.MsgTypePDUSessionModificationComplete)
	return encodeWithBuilder(msg.EncodePDUSessionModificationComplete)
}

type PduSessionReleaseRequest struct {
	PduSessionID byte
	Pti          byte
}

func (m *PduSessionReleaseRequest) Encode() *utils.Buffer {
	msg := extmsg.NewPDUSessionReleaseRequest(extnas.MsgTypePDUSessionReleaseRequest)
	msg.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(PD_5G_SESSION_MANAGEMENT)
	msg.PDUSessionID.SetPDUSessionID(m.PduSessionID)
	msg.PTI.SetPTI(m.Pti)
	msg.PDUSESSIONRELEASEREQUESTMessageIdentity.SetMessageType(extnas.MsgTypePDUSessionReleaseRequest)
	return encodeWithBuilder(msg.EncodePDUSessionReleaseRequest)
}

type PduSessionReleaseCommand struct {
	PduSessionID byte
	Pti          byte
}

func DecodePduSessionReleaseCommand(data []byte) (*PduSessionReleaseCommand, error) {
	wire := append([]byte(nil), data...)
	msg := extnas.NewMessage()
	if err := msg.PlainNasDecode(&wire); err != nil {
		return nil, err
	}
	if msg.GsmMessage == nil || msg.GsmMessage.PDUSessionReleaseCommand == nil {
		return nil, fmt.Errorf("not a PDU Session Release Command")
	}
	src := msg.GsmMessage.PDUSessionReleaseCommand
	return &PduSessionReleaseCommand{
		PduSessionID: src.PDUSessionID.GetPDUSessionID(),
		Pti:          src.PTI.GetPTI(),
	}, nil
}

type PduSessionReleaseComplete struct {
	PduSessionID byte
	Pti          byte
}

func (m *PduSessionReleaseComplete) Encode() *utils.Buffer {
	msg := extmsg.NewPDUSessionReleaseComplete(extnas.MsgTypePDUSessionReleaseComplete)
	msg.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(PD_5G_SESSION_MANAGEMENT)
	msg.PDUSessionID.SetPDUSessionID(m.PduSessionID)
	msg.PTI.SetPTI(m.Pti)
	msg.PDUSESSIONRELEASECOMPLETEMessageIdentity.SetMessageType(extnas.MsgTypePDUSessionReleaseComplete)
	return encodeWithBuilder(msg.EncodePDUSessionReleaseComplete)
}
