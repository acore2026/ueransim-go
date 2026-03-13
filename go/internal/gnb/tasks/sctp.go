package tasks

import (
	"context"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/lib/sctp"
)

type GnbSctpTaskHandler struct {
	logger      logging.Logger
	amfAddr     string
	amfPort     int
	client      *sctp.ClientTaskHandler
	ngapTask    *runtime.Task
}

func NewGnbSctpTaskHandler(logger logging.Logger, amfAddr string, amfPort int, ngapTask *runtime.Task) *GnbSctpTaskHandler {
	return &GnbSctpTaskHandler{
		logger:   logger.With("component", "sctp"),
		amfAddr:  amfAddr,
		amfPort:  amfPort,
		ngapTask: ngapTask,
	}
}

func (h *GnbSctpTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.logger.Info("starting SCTP task")
	h.client = sctp.NewClientTaskHandler("", h.amfAddr, h.amfPort, t, h.logger)
	return h.client.OnStart(ctx, t)
}

func (h *GnbSctpTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case sctp.MessageTypeSctpReceive:
		h.logger.Info("received SCTP message from AMF, forwarding to NGAP")
		return h.ngapTask.Send(msg)
	case "ngap_to_sctp":
		h.logger.Info("forwarding NGAP message to AMF via SCTP")
		payload := msg.Payload.(sctp.SendMessage)
		return h.client.OnMessage(ctx, runtime.Message{
			Type:    sctp.MessageTypeSctpSend,
			Payload: payload,
		})
	}
	return nil
}

func (h *GnbSctpTaskHandler) OnStop(ctx context.Context) error {
	return h.client.OnStop(ctx)
}
