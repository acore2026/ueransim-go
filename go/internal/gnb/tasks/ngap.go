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
	rlsTask  *runtime.Task
}

func NewGnbNgapTaskHandler(logger logging.Logger, gnbName string, gnbId []byte, plmnId []byte, sctpTask *runtime.Task, rlsTask *runtime.Task) *GnbNgapTaskHandler {
	return &GnbNgapTaskHandler{
		logger:   logger.With("component", "ngap"),
		gnbName:  gnbName,
		gnbId:    gnbId,
		plmnId:   plmnId,
		sctpTask: sctpTask,
		rlsTask:  rlsTask,
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
			Ppid:   60, // NGAP PPID
			Data:   encoded,
		},
	})
}

func (h *GnbNgapTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case "rls_to_ngap":
		nasPdu := msg.Payload.([]byte)
		h.logger.Info("received NAS PDU from RLS, sending InitialUEMessage")
		
		// For InitialUEMessage, we need a RAN UE NGAP ID
		// In a real gNB, this would be managed per UE context.
		// For testing, let's use a fixed ID.
		ranUeNgapId := int64(1)
		
		// Build User Location Info (dummy values)
		tac := []byte{0x00, 0x00, 0x01}
		nrCellId := []byte{0x00, 0x00, 0x00, 0x00, 0x10}
		uli := ngap.BuildUserLocationInformationNR(h.plmnId, tac, nrCellId)
		
		pdu, err := ngap.BuildInitialUEMessage(ranUeNgapId, nasPdu, uli)
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
				Ppid:   60,
				Data:   encoded,
			},
		})
		
	case sctp.MessageTypeSctpReceive:
		h.logger.Info("received NGAP message from AMF")
		
		data := msg.Payload.(sctp.ReceiveMessage).Data
		pdu, err := ngap.Decode(data)
		if err != nil {
			h.logger.Error("failed to decode NGAP message", "error", err)
			return nil
		}
		
		nasPdu := ngap.GetNasPdu(pdu)
		if nasPdu != nil {
			h.logger.Info("forwarding downlink NAS PDU to RLS")
			return h.rlsTask.Send(runtime.Message{
				Type:    "ngap_to_rls",
				Payload: nasPdu,
			})
		}
	}
	return nil
}

func (h *GnbNgapTaskHandler) OnStop(ctx context.Context) error {
	return nil
}
