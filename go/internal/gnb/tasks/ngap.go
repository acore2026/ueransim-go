package tasks

import (
	"context"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/lib/sctp"
	"github.com/acore2026/ueransim-go/internal/ngap"
	"github.com/free5gc/ngap/ngapType"
)

type GnbNgapTaskHandler struct {
	logger   logging.Logger
	gnbName  string
	gnbId    []byte
	plmnId   []byte
	uli      *ngapType.UserLocationInformation
	sctpTask *runtime.Task
	rlsTask  *runtime.Task

	// For testing, track if we already sent InitialUEMessage
	initialSent bool
	amfUeId     int64
}

func NewGnbNgapTaskHandler(logger logging.Logger, gnbName string, gnbId []byte, plmnId []byte, uli *ngapType.UserLocationInformation, sctpTask *runtime.Task, rlsTask *runtime.Task) *GnbNgapTaskHandler {
	return &GnbNgapTaskHandler{
		logger:      logger.With("component", "ngap"),
		gnbName:     gnbName,
		gnbId:       gnbId,
		plmnId:      plmnId,
		uli:         uli,
		sctpTask:    sctpTask,
		rlsTask:     rlsTask,
		initialSent: false,
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

		var pdu *ngapType.NGAPPDU
		var err error

		if !h.initialSent {
			h.logger.Info("received NAS PDU from RLS, sending InitialUEMessage")
			ranUeNgapId := int64(1)
			pdu, err = ngap.BuildInitialUEMessage(ranUeNgapId, nasPdu, h.uli)
			h.initialSent = true
		} else {
			h.logger.Info("received NAS PDU from RLS, sending UplinkNASTransport")
			ranUeNgapId := int64(1)
			pdu, err = ngap.BuildUplinkNASTransport(ranUeNgapId, h.amfUeId, nasPdu, h.uli)
		}

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

		// Update AMF UE ID if present
		if pdu.Present == ngapType.NGAPPDUPresentInitiatingMessage {
			ini := pdu.InitiatingMessage
			if ini.ProcedureCode.Value == ngapType.ProcedureCodeDownlinkNASTransport {
				down := ini.Value.DownlinkNASTransport
				for _, ie := range down.ProtocolIEs.List {
					if ie.Id.Value == ngapType.ProtocolIEIDAMFUENGAPID {
						h.amfUeId = ie.Value.AMFUENGAPID.Value
						h.logger.Info("updated AMF UE NGAP ID", "id", h.amfUeId)
					}
				}
			}
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
