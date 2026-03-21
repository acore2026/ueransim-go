package gnb

import (
	"context"
	"net"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/gnb/tasks"
	"github.com/acore2026/ueransim-go/internal/gnbctx"
	"github.com/acore2026/ueransim-go/internal/lib/udp"
	"github.com/acore2026/ueransim-go/internal/rls"
	"github.com/acore2026/ueransim-go/internal/rrc"
)

type RlsTaskHandler struct {
	logger     logging.Logger
	udpHandler *udp.ServerTaskHandler
	addr       string
	ngapTask   *runtime.Task
	gtpTask    *runtime.Task
	lastUeAddr *net.UDPAddr
}

func NewRlsTaskHandler(logger logging.Logger, addr string, ngapTask *runtime.Task, gtpTask *runtime.Task) *RlsTaskHandler {
	return &RlsTaskHandler{
		logger:   logger.With("component", "rls"),
		addr:     addr,
		ngapTask: ngapTask,
		gtpTask:  gtpTask,
	}
}

func (h *RlsTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.logger.Info("RLS task started", "addr", h.addr)

	udpAddr, err := net.ResolveUDPAddr("udp", h.addr)
	if err != nil {
		return err
	}

	h.udpHandler = udp.NewServerTaskHandler(udpAddr, t, h.logger)
	return h.udpHandler.OnStart(ctx, t)
}

func (h *RlsTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case udp.MessageTypeUdpReceive:
		h.logger.Info("received radio packet from UE")

		payload := msg.Payload.(udp.ReceiveMessage)
		h.lastUeAddr = payload.From

		rlsMsg, err := rls.Decode(payload.Data)
		if err != nil {
			h.logger.Error("failed to decode RLS message", "error", err)
			return nil
		}

		h.logger.Info("decoded RLS message", "type", rlsMsg.MsgType, "pduType", rlsMsg.PduType)

		if rlsMsg.MsgType == rls.PDU_TRANSMISSION && rlsMsg.PduType == rls.PDU_TYPE_RRC {
			// Extract NAS PDU from authentic bit-accurate RRC messages.
			nasPdu := []byte(nil)

			if len(rlsMsg.Pdu) > 2 {
				// Detect authentic bit-stream
				// RRCSetupComplete:
				// Octet 0: 0 (c1) | 0010 (index 2) | 00 (trans 0) | 0 (critical 0) = 0x10
				// Octet 1: 0000 (plmn 1) | 0 (reg 0) | 0 (guami 0) | 0 (nssai 0) | L (len bit 0) = 0x00 or 0x01
				// Octet 2: LLLLLLL (len bits 1-7) | N (nas bit 0)
				
				// ULInformationTransfer:
				// Octet 0: 0 (c1) | 0111 (index 7) | 0 (critical 0) | LL (len bits 0-1) = 0x38, 0x39, 0x3A, 0x3B
				// Octet 1: LLLLLL (len bits 2-7) | NN (nas bits 0-1)

				if rlsMsg.Pdu[0] == 0x10 && (rlsMsg.Pdu[1]&0xFE) == 0x00 {
					// RRCSetupComplete
					nasLen := int(rlsMsg.Pdu[1]&0x01)<<7 | int(rlsMsg.Pdu[2]>>1)
					if len(rlsMsg.Pdu) >= 3+nasLen {
						nasPdu = make([]byte, nasLen)
						for i := 0; i < nasLen; i++ {
							nasPdu[i] = uint8(rlsMsg.Pdu[i+2]&0x01)<<7 | uint8(rlsMsg.Pdu[i+3]>>1)
						}
					}
				} else if (rlsMsg.Pdu[0] & 0xFC) == 0x38 {
					// ULInformationTransfer
					nasLen := int(rlsMsg.Pdu[0]&0x03)<<6 | int(rlsMsg.Pdu[1]>>2)
					if len(rlsMsg.Pdu) >= 2+nasLen {
						nasPdu = make([]byte, nasLen)
						for i := 0; i < nasLen; i++ {
							nasPdu[i] = uint8(rlsMsg.Pdu[i+1]&0x03)<<6 | uint8(rlsMsg.Pdu[i+2]>>2)
						}
					}
				}
			}

			if nasPdu != nil {
				return h.ngapTask.Send(runtime.Message{
					Type:    "rls_to_ngap",
					Payload: nasPdu,
				})
			}
		}
		if rlsMsg.MsgType == rls.PDU_TRANSMISSION && rlsMsg.PduType == rls.PduTypeData {
			return h.gtpTask.Send(runtime.Message{
				Type: tasks.MessageTypeRlsToGtp,
				Payload: gnbctx.UplinkPacket{
					SessionID: uint8(rlsMsg.Payload),
					Data:      append([]byte(nil), rlsMsg.Pdu...),
				},
			})
		}

	case "ngap_to_rls":
		if h.lastUeAddr == nil {
			h.logger.Info("cannot send downlink message: no UE address known")
			return nil
		}

		nasPdu := msg.Payload.([]byte)
		h.logger.Info("received NAS from NGAP, sending to UE via RLS")

		// Wrap NAS in authentic DLInformationTransfer
		rrcPdu := rrc.BuildDLInformationTransfer(nasPdu)

		rlsMsg := &rls.RlsMessage{
			MsgType: rls.PDU_TRANSMISSION,
			Sti:     1,
			PduType: rls.PDU_TYPE_RRC,
			Pdu:     rrcPdu,
		}

		encoded, err := rlsMsg.Encode()
		if err != nil {
			return err
		}

		return h.udpHandler.Send(h.lastUeAddr, encoded)
	case tasks.MessageTypeGtpToRls:
		if h.lastUeAddr == nil {
			h.logger.Info("cannot send downlink user-plane packet: no UE address known")
			return nil
		}
		payload := msg.Payload.(gnbctx.DownlinkPacket)
		rlsMsg := &rls.RlsMessage{
			MsgType: rls.PDU_TRANSMISSION,
			Sti:     1,
			PduType: rls.PduTypeData,
			Payload: uint32(payload.SessionID),
			Pdu:     payload.Data,
		}
		encoded, err := rlsMsg.Encode()
		if err != nil {
			return err
		}
		return h.udpHandler.Send(h.lastUeAddr, encoded)
	}
	return nil
}

func (h *RlsTaskHandler) OnStop(ctx context.Context) error {
	return h.udpHandler.OnStop(ctx)
}
