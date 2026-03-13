package ue

import (
	"context"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/rrc"
)

type RrcTaskHandler struct {
	logger  logging.Logger
	rlsTask *runtime.Task
}

func NewRrcTaskHandler(logger logging.Logger, rlsTask *runtime.Task) *RrcTaskHandler {
	return &RrcTaskHandler{
		logger:  logger.With("component", "rrc"),
		rlsTask: rlsTask,
	}
}

func (h *RrcTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.logger.Info("RRC task started")
	return nil
}

func (h *RrcTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case "nas_to_rrc":
		nasPdu := msg.Payload.([]byte)
		h.logger.Info("received NAS PDU, wrapping in RRC")
		
		// For the very first message, we usually send RRCSetupRequest (which doesn't carry NAS)
		// then RRCSetupComplete (which carries NAS).
		// Simplified: just wrap NAS in RRCSetupComplete and send over RLS.
		
		rrcPdu := rrc.BuildRRCSetupComplete(nasPdu)
		
		return h.rlsTask.Send(runtime.Message{
			Type: "rrc_to_rls",
			Payload: rrcPdu,
		})
	}
	return nil
}

func (h *RrcTaskHandler) OnStop(context.Context) error {
	return nil
}
