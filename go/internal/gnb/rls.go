package gnb

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
	udpHandler *udp.ServerTaskHandler
	addr       string
}

func NewRlsTaskHandler(logger logging.Logger, addr string) *RlsTaskHandler {
	return &RlsTaskHandler{
		logger: logger.With("component", "rls"),
		addr:   addr,
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
		rlsMsg, err := rls.Decode(payload.Data)
		if err != nil {
			h.logger.Error("failed to decode RLS message", "error", err)
			return nil
		}
		
		h.logger.Info("decoded RLS message", "type", rlsMsg.MsgType, "pduType", rlsMsg.PduType)
		// TODO: Forward to RRC
	}
	return nil
}

func (h *RlsTaskHandler) OnStop(ctx context.Context) error {
	return h.udpHandler.OnStop(ctx)
}
