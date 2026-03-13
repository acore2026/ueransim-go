package tasks

import (
	"context"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/lib/sctp"
	"github.com/acore2026/ueransim-go/internal/ngap"
)

type GnbNgapTaskHandler struct {
	logger   logging.Logger
	gnbName  string
	gnbId    []byte
	plmnId   []byte
	sctpTask *runtime.Task
}

func NewGnbNgapTaskHandler(logger logging.Logger, gnbName string, gnbId []byte, plmnId []byte, sctpTask *runtime.Task) *GnbNgapTaskHandler {
	return &GnbNgapTaskHandler{
		logger:   logger.With("component", "ngap"),
		gnbName:  gnbName,
		gnbId:    gnbId,
		plmnId:   plmnId,
		sctpTask: sctpTask,
	}
}

func (h *GnbNgapTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.logger.Info("NGAP task started, sending NGSetupRequest")
	
	pdu, err := ngap.BuildNGSetupRequest(h.gnbName, h.gnbId, 24, h.plmnId)
	if err != nil {
		return err
	}
	
	encoded, err := ngap.Encode(pdu)
	if err != nil {
		return err
	}
	
	return h.sctpTask.Send(runtime.Message{
		Type: "ngap_to_sctp",
		Payload: sctp.SendMessage{
			Stream: 0,
			Ppid:   0x3c000000, // NGAP PPID
			Data:   encoded,
		},
	})
}

func (h *GnbNgapTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case sctp.MessageTypeSctpReceive:
		h.logger.Info("received NGAP message from AMF")
		// TODO: Decode and process NGAP PDU
	}
	return nil
}

func (h *GnbNgapTaskHandler) OnStop(ctx context.Context) error {
	return nil
}
