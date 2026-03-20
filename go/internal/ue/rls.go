package ue

import (
	"context"
	"net"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/lib/udp"
	"github.com/acore2026/ueransim-go/internal/rls"
	"github.com/acore2026/ueransim-go/internal/ue/tun"
)

type RlsTaskHandler struct {
	logger       logging.Logger
	gnbAddr      *net.UDPAddr
	udpHandler   *udp.ServerTaskHandler
	sti          uint64
	rrcTask      *runtime.Task
	tunTask      *runtime.Task
	sessionID    byte
	sessionReady bool
}

func NewRlsTaskHandler(logger logging.Logger, gnbAddr string, sti uint64, rrcTask *runtime.Task, tunTask *runtime.Task) (*RlsTaskHandler, error) {
	addr, err := net.ResolveUDPAddr("udp", gnbAddr)
	if err != nil {
		return nil, err
	}

	return &RlsTaskHandler{
		logger:  logger.With("component", "rls"),
		gnbAddr: addr,
		sti:     sti,
		rrcTask: rrcTask,
		tunTask: tunTask,
	}, nil
}

func (h *RlsTaskHandler) SetRrcTask(t *runtime.Task) {
	h.rrcTask = t
}

func (h *RlsTaskHandler) SetTunTask(t *runtime.Task) {
	h.tunTask = t
}

func (h *RlsTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.logger.Info("RLS task started")

	// Start an ephemeral UDP server to talk to gNB
	// Pass the RLS task itself as the target for UDP receive messages
	h.udpHandler = udp.NewServerTaskHandler(nil, t, h.logger)
	return h.udpHandler.OnStart(ctx, t)
}

func (h *RlsTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case "rrc_to_rls":
		rrcPdu := msg.Payload.([]byte)

		rlsMsg := &rls.RlsMessage{
			MsgType: rls.PDU_TRANSMISSION,
			Sti:     h.sti,
			PduType: rls.PDU_TYPE_RRC,
			Pdu:     rrcPdu,
		}

		encoded, err := rlsMsg.Encode()
		if err != nil {
			return err
		}

		return h.udpHandler.Send(h.gnbAddr, encoded)
	case tun.MessageTypeTunToApp:
		if !h.sessionReady {
			h.logger.Info("dropping UE data packet before session is ready")
			return nil
		}
		payload := msg.Payload.(tun.TunToAppMessage)
		rlsMsg := &rls.RlsMessage{
			MsgType: rls.PDU_TRANSMISSION,
			Sti:     h.sti,
			PduType: rls.PduTypeData,
			Payload: uint32(payload.Psi),
			Pdu:     payload.Data,
		}
		encoded, err := rlsMsg.Encode()
		if err != nil {
			return err
		}
		return h.udpHandler.Send(h.gnbAddr, encoded)
	case "nas_session_ready":
		h.sessionID = msg.Payload.(byte)
		h.sessionReady = true
		h.logger.Info("user-plane session marked ready", "sessionID", h.sessionID)

	case udp.MessageTypeUdpReceive:
		// Handle incoming radio packets from gNB
		h.logger.Info("received radio packet from gNB")

		payload := msg.Payload.(udp.ReceiveMessage)
		rlsMsg, err := rls.Decode(payload.Data)
		if err != nil {
			h.logger.Error("failed to decode RLS message", "error", err)
			return nil
		}

		if rlsMsg.MsgType == rls.PDU_TRANSMISSION && rlsMsg.PduType == rls.PDU_TYPE_RRC {
			h.logger.Info("forwarding RRC PDU to RRC task")
			return h.rrcTask.Send(runtime.Message{
				Type:    "rls_to_rrc",
				Payload: rlsMsg.Pdu,
			})
		}
		if rlsMsg.MsgType == rls.PDU_TRANSMISSION && rlsMsg.PduType == rls.PduTypeData && h.tunTask != nil {
			h.logger.Info("forwarding user-plane packet to TUN", "sessionID", rlsMsg.Payload, "len", len(rlsMsg.Pdu))
			return h.tunTask.Send(runtime.Message{
				Type: tun.MessageTypeAppToTun,
				Payload: tun.AppToTunMessage{
					Data: rlsMsg.Pdu,
				},
			})
		}
	}
	return nil
}

func (h *RlsTaskHandler) OnStop(ctx context.Context) error {
	return h.udpHandler.OnStop(ctx)
}
