package gnb

import (
	"context"
	"encoding/hex"
	"net"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/gnb/tasks"
	"github.com/acore2026/ueransim-go/internal/lib/udp"
	"github.com/acore2026/ueransim-go/internal/rlc"
	"github.com/acore2026/ueransim-go/internal/rls"
	"github.com/acore2026/ueransim-go/internal/rrc"
)

type RlsTaskHandler struct {
	logger     logging.Logger
	udpHandler *udp.ServerTaskHandler
	addr       string
	rlcTask    *runtime.Task
	ngapTask   *runtime.Task
	gtpTask    *runtime.Task
	lastUeAddr *net.UDPAddr
}

func NewRlsTaskHandler(logger logging.Logger, addr string, rlcTask *runtime.Task, gtpTask *runtime.Task) *RlsTaskHandler {
	return &RlsTaskHandler{
		logger:  logger.With("component", "rls"),
		addr:    addr,
		rlcTask: rlcTask,
		gtpTask: gtpTask,
	}
}

func (h *RlsTaskHandler) SetRrcTask(t *runtime.Task) {
	h.rlcTask = t
}

func (h *RlsTaskHandler) SetNgapTask(t *runtime.Task) {
	h.ngapTask = t
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
	case "rlc_to_rls":
		payload := msg.Payload.(rlc.RlcToRlsMessage)
		if h.lastUeAddr == nil {
			return nil
		}

		rlsMsg := &rls.RlsMessage{
			MsgType: rls.PDU_TRANSMISSION,
			Sti:     1,
			PduType: payload.PduType,
			Pdu:     payload.Pdu,
			Payload: payload.Payload,
		}

		encoded, err := rlsMsg.Encode()
		if err != nil {
			return err
		}

		return h.udpHandler.Send(h.lastUeAddr, encoded)

	case "rlc_to_ccch":
		// UL-CCCH Message index 0 is RRCSetupRequest
		h.logger.Info("received RRCSetupRequest, sending RRCSetup")
		rrcResp := rrc.BuildRRCSetup()
		return h.rlcTask.Send(runtime.Message{
			Type: "upper_to_rlc",
			Payload: rlc.UpperToRlcMessage{
				Mode: rlc.ModeTM,
				Pdu:  rrcResp,
			},
		})

	case "rlc_to_nas":
		nasPdu := msg.Payload.([]byte)
		h.logger.Info("extracted NAS PDU from RLC", "hex", hex.EncodeToString(nasPdu))
		if h.ngapTask != nil {
			return h.ngapTask.Send(runtime.Message{
				Type:    "rls_to_ngap",
				Payload: nasPdu,
			})
		}

	case tasks.MessageTypeRlsToGtp:
		return h.gtpTask.Send(msg)

	case udp.MessageTypeUdpReceive:
		h.logger.Info("received radio packet from UE")

		payload := msg.Payload.(udp.ReceiveMessage)
		h.lastUeAddr = payload.From

		rlsMsg, err := rls.Decode(payload.Data)
		if err != nil {
			h.logger.Error("failed to decode RLS message", "error", err)
			return nil
		}

		h.logger.Info("decoded RLS message", "type", rlsMsg.MsgType, "pduType", rlsMsg.PduType, "payload", rlsMsg.Payload)

		if rlsMsg.MsgType == rls.PDU_TRANSMISSION {
			if h.rlcTask == nil {
				h.logger.Warn("RLC task not set, dropping PDU")
				return nil
			}
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
