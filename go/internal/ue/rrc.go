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
	nasTask *runtime.Task
}

func NewRrcTaskHandler(logger logging.Logger, rlsTask *runtime.Task, nasTask *runtime.Task) *RrcTaskHandler {
	return &RrcTaskHandler{
		logger:  logger.With("component", "rrc"),
		rlsTask: rlsTask,
		nasTask: nasTask,
	}
}

func (h *RrcTaskHandler) SetNasTask(t *runtime.Task) {
	h.nasTask = t
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
		
	case "rls_to_rrc":
		h.logger.Info("received RRC PDU from RLS")
		
		rrcPdu := msg.Payload.([]byte)
		// Extract NAS PDU from simplified RRC container
		if len(rrcPdu) > 5 && rrcPdu[0] == 0x01 {
			nasLen := int(rrcPdu[1])<<24 | int(rrcPdu[2])<<16 | int(rrcPdu[3])<<8 | int(rrcPdu[4])
			if len(rrcPdu) >= 5+nasLen {
				nasPdu := rrcPdu[5 : 5+nasLen]
				h.logger.Info("forwarding NAS PDU to NAS task")
				return h.nasTask.Send(runtime.Message{
					Type: "rrc_to_nas",
					Payload: nasPdu,
				})
			}
		}
	}
	return nil
}

func (h *RrcTaskHandler) OnStop(context.Context) error {
	return nil
}
