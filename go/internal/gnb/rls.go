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
			// Extract NAS PDU from simplified RRC message (the container)
			if len(rlsMsg.Pdu) > 5 && rlsMsg.Pdu[0] == 0x01 {
				// We'll just pass the whole RRC for now or extract the NAS.
				// For InitialUEMessage, we need the NAS PDU.

				nasLen := int(rlsMsg.Pdu[1])<<24 | int(rlsMsg.Pdu[2])<<16 | int(rlsMsg.Pdu[3])<<8 | int(rlsMsg.Pdu[4])
				if len(rlsMsg.Pdu) >= 5+nasLen {
					nasPdu := rlsMsg.Pdu[5 : 5+nasLen]

					return h.ngapTask.Send(runtime.Message{
						Type:    "rls_to_ngap",
						Payload: nasPdu,
					})
				}
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

		// Wrap NAS in simplified RRC and then in RLS
		rrcPdu := rls.BuildSimpleRrc(nasPdu)

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
