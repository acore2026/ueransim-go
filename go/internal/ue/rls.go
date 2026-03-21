package ue

import (
	"context"
	"net"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/lib/udp"
	"github.com/acore2026/ueransim-go/internal/rlc"
	"github.com/acore2026/ueransim-go/internal/rls"
	"github.com/acore2026/ueransim-go/internal/ue/tun"
)

type RlsTaskHandler struct {
	logger       logging.Logger
	gnbAddr      *net.UDPAddr
	udpHandler   *udp.ServerTaskHandler
	sti          uint64
	rlcTask      *runtime.Task
	tunTask      *runtime.Task
	sessionID    byte
	sessionReady bool
}

func NewRlsTaskHandler(logger logging.Logger, gnbAddr string, sti uint64, rlcTask *runtime.Task, tunTask *runtime.Task) (*RlsTaskHandler, error) {
	addr, err := net.ResolveUDPAddr("udp", gnbAddr)
	if err != nil {
		return nil, err
	}

	return &RlsTaskHandler{
		logger:  logger.With("component", "rls"),
		gnbAddr: addr,
		sti:     sti,
		rlcTask: rlcTask,
		tunTask: tunTask,
	}, nil
}

func (h *RlsTaskHandler) SetRrcTask(t *runtime.Task) {
	h.rlcTask = t
}

func (h *RlsTaskHandler) SetTunTask(t *runtime.Task) {
	h.tunTask = t
}

func (h *RlsTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.logger.Info("RLS task started")

	h.udpHandler = udp.NewServerTaskHandler(nil, t, h.logger)
	return h.udpHandler.OnStart(ctx, t)
}

func (h *RlsTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case "rlc_to_rls":
		payload := msg.Payload.(rlc.RlcToRlsMessage)

		rlsMsg := &rls.RlsMessage{
			MsgType: rls.PDU_TRANSMISSION,
			Sti:     h.sti,
			PduType: payload.PduType,
			Pdu:     payload.Pdu,
			Payload: payload.Payload,
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
		// Send to RLC for UM framing
		if h.rlcTask != nil {
			return h.rlcTask.Send(runtime.Message{
				Type: "upper_to_rlc",
				Payload: rlc.UpperToRlcMessage{
					Mode: rlc.ModeUM,
					Pdu:  payload.Data,
				},
			})
		}

	case "nas_session_ready":
		h.sessionID = msg.Payload.(byte)
		h.sessionReady = true
		h.logger.Info("user-plane session marked ready", "sessionID", h.sessionID)

	case udp.MessageTypeUdpReceive:
		h.logger.Info("received radio packet from gNB")

		payload := msg.Payload.(udp.ReceiveMessage)
		rlsMsg, err := rls.Decode(payload.Data)
		if err != nil {
			h.logger.Error("failed to decode RLS message", "error", err)
			return nil
		}

		if rlsMsg.MsgType == rls.PDU_TRANSMISSION {
			h.logger.Info("forwarding PDU to RLC task")
			return h.rlcTask.Send(runtime.Message{
				Type: "rls_to_rlc",
				Payload: rlc.RlsToRlcMessage{
					PduType: rlsMsg.PduType,
					Pdu:     rlsMsg.Pdu,
					Payload: rlsMsg.Payload,
				},
			})
		}
	}
	return nil
}

func (h *RlsTaskHandler) OnStop(ctx context.Context) error {
	return h.udpHandler.OnStop(ctx)
}
