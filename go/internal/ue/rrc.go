package ue

import (
	"context"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/rrc"
)

type RrcTaskHandler struct {
	logger      logging.Logger
	rlsTask     *runtime.Task
	nasTask     *runtime.Task
	isFirstResp bool
}

func NewRrcTaskHandler(logger logging.Logger, rlsTask *runtime.Task, nasTask *runtime.Task) *RrcTaskHandler {
	return &RrcTaskHandler{
		logger:      logger.With("component", "rrc"),
		rlsTask:     rlsTask,
		nasTask:     nasTask,
		isFirstResp: true,
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

		var rrcPdu []byte
		if h.isFirstResp {
			h.logger.Info("using RRCSetupComplete for initial registration")
			rrcPdu = rrc.BuildRRCSetupComplete(nasPdu)
			h.isFirstResp = false
		} else {
			h.logger.Info("using ULInformationTransfer for subsequent NAS")
			rrcPdu = rrc.BuildULInformationTransfer(nasPdu)
		}

		return h.rlsTask.Send(runtime.Message{
			Type:    "rrc_to_rls",
			Payload: rrcPdu,
		})

	case "rls_to_rrc":
		h.logger.Info("received RRC PDU from RLS")
		rrcPdu := msg.Payload.([]byte)

		// Authentic bit-stream decoding:
		// DLInformationTransfer:
		// Octet 0: 0 (c1) | 0101 (index 5) | 00 (trans 0) | 0 (critical 0) = 0x28
		// Octet 1: length (8 bits)
		// Octet 2...: nasPdu

		nasPdu := []byte(nil)
		if len(rrcPdu) > 2 && rrcPdu[0] == 0x28 {
			nasLen := int(rrcPdu[1])
			if len(rrcPdu) >= 2+nasLen {
				nasPdu = rrcPdu[2 : 2+nasLen]
			}
		}

		if nasPdu != nil {
			h.logger.Info("forwarding NAS PDU to NAS task")
			return h.nasTask.Send(runtime.Message{
				Type:    "rrc_to_nas",
				Payload: nasPdu,
			})
		}
	}

	return nil
}

func (h *RrcTaskHandler) OnStop(context.Context) error {
	return nil
}
