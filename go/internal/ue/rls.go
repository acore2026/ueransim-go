package ue

import (
	"context"
	"net"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/lib/udp"
	"github.com/acore2026/ueransim-go/internal/rls"
)

type RlsTaskHandler struct {
	logger     logging.Logger
	gnbAddr    *net.UDPAddr
	udpHandler *udp.ServerTaskHandler
	sti        uint64
	rrcTask    *runtime.Task
}

func NewRlsTaskHandler(logger logging.Logger, gnbAddr string, sti uint64, rrcTask *runtime.Task) (*RlsTaskHandler, error) {
	addr, err := net.ResolveUDPAddr("udp", gnbAddr)
	if err != nil {
		return nil, err
	}
	
	return &RlsTaskHandler{
		logger:  logger.With("component", "rls"),
		gnbAddr: addr,
		sti:     sti,
		rrcTask: rrcTask,
	}, nil
}

func (h *RlsTaskHandler) SetRrcTask(t *runtime.Task) {
	h.rrcTask = t
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
				Type: "rls_to_rrc",
				Payload: rlsMsg.Pdu,
			})
		}
	}
	return nil
}

func (h *RlsTaskHandler) OnStop(ctx context.Context) error {
	return h.udpHandler.OnStop(ctx)
}
