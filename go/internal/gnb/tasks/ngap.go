package tasks

import (
	"context"
	"encoding/hex"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/gnbctx"
	"github.com/acore2026/ueransim-go/internal/lib/sctp"
	"github.com/acore2026/ueransim-go/internal/ngap"
	"github.com/acore2026/ueransim-go/internal/rlc"
	"github.com/acore2026/ueransim-go/internal/rrc"
	"github.com/free5gc/ngap/ngapType"
)

type GnbNgapTaskHandler struct {
	logger       logging.Logger
	gnbName      string
	gnbId        []byte
	plmnId       []byte
	uli          *ngapType.UserLocationInformation
	localGTPIP   string
	sctpTask     *runtime.Task
	rlcTask      *runtime.Task
	sessionStore *gnbctx.SessionStore

	// For testing, track if we already sent InitialUEMessage
	initialSent bool
	amfUeId     int64
}

func NewGnbNgapTaskHandler(logger logging.Logger, gnbName string, gnbId []byte, plmnId []byte, uli *ngapType.UserLocationInformation, localGTPIP string, sctpTask *runtime.Task, rlcTask *runtime.Task, sessionStore *gnbctx.SessionStore) *GnbNgapTaskHandler {
	return &GnbNgapTaskHandler{
		logger:       logger.With("component", "ngap"),
		gnbName:      gnbName,
		gnbId:        gnbId,
		plmnId:       plmnId,
		uli:          uli,
		localGTPIP:   localGTPIP,
		sctpTask:     sctpTask,
		rlcTask:      rlcTask,
		sessionStore: sessionStore,
		initialSent:  false,
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
			h.logger.Info("received NAS PDU from RLS, sending InitialUEMessage", "hex", hex.EncodeToString(nasPdu))
			ranUeNgapId := int64(1)
			pdu, err = ngap.BuildInitialUEMessage(ranUeNgapId, nasPdu, h.uli)
			h.initialSent = true
		} else {
			h.logger.Info("received NAS PDU from RLS, sending UplinkNASTransport", "hex", hex.EncodeToString(nasPdu))
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
		data := msg.Payload.(sctp.ReceiveMessage).Data
		pdu, err := ngap.Decode(data)
		if err != nil {
			h.logger.Error("failed to decode NGAP message", "error", err)
			return nil
		}

		if pdu.Present == ngapType.NGAPPDUPresentInitiatingMessage {
			h.logger.Info("received NGAP initiating message", "procedureCode", pdu.InitiatingMessage.ProcedureCode.Value)
		} else if pdu.Present == ngapType.NGAPPDUPresentSuccessfulOutcome {
			h.logger.Info("received NGAP successful outcome", "procedureCode", pdu.SuccessfulOutcome.ProcedureCode.Value)
		}

		h.captureAmfUeID(pdu)

		if ctxData, err := ngap.ParseInitialContextSetupRequest(pdu); err != nil {
			h.logger.Error("failed to parse InitialContextSetupRequest", "error", err)
			return nil
		} else if ctxData != nil {
			h.amfUeId = ctxData.AMFUENGAPID
			responses := make([]ngap.SessionResourceSetupResponse, 0, len(ctxData.Sessions))
			for _, session := range ctxData.Sessions {
				record := h.sessionStore.Upsert(gnbctx.SessionSetupRequest{
					RANUENGAPID: ctxData.RANUENGAPID,
					AMFUENGAPID: ctxData.AMFUENGAPID,
					SessionID:   session.PDUSessionID,
					RemoteIP:    session.RemoteGTPIP,
					RemoteTEID:  session.RemoteTEID,
					QFIs:        session.QFIs,
				}, h.localGTPIP)
				responses = append(responses, ngap.SessionResourceSetupResponse{
					PDUSessionID: record.SessionID,
					LocalGTPIP:   record.LocalIP,
					LocalTEID:    record.LocalTEID,
					QFIs:         record.QFIs,
				})
			}
			responsePDU, err := ngap.BuildInitialContextSetupResponse(ctxData.AMFUENGAPID, ctxData.RANUENGAPID, responses)
			if err != nil {
				return err
			}
			encoded, err := ngap.Encode(responsePDU)
			if err != nil {
				return err
			}
			if err := h.sctpTask.Send(runtime.Message{
				Type: "ngap_to_sctp",
				Payload: sctp.SendMessage{
					Stream: 0,
					Ppid:   60,
					Data:   encoded,
				},
			}); err != nil {
				return err
			}
			h.logger.Info("sent InitialContextSetupResponse", "sessions", len(responses))
		}

		nasPdu := ngap.GetNasPdu(pdu)
		if nasPdu != nil {
			h.logger.Info("forwarding downlink NAS PDU to RLC")
			rrcPdu := rrc.BuildDLInformationTransfer(nasPdu)
			return h.rlcTask.Send(runtime.Message{
				Type: "upper_to_rlc",
				Payload: rlc.UpperToRlcMessage{
					Mode: rlc.ModeUM,
					Pdu:  rrcPdu,
				},
			})
		}
	}
	return nil
}

func (h *GnbNgapTaskHandler) captureAmfUeID(pdu *ngapType.NGAPPDU) {
	if pdu.Present != ngapType.NGAPPDUPresentInitiatingMessage {
		return
	}
	ini := pdu.InitiatingMessage
	switch ini.ProcedureCode.Value {
	case ngapType.ProcedureCodeDownlinkNASTransport:
		down := ini.Value.DownlinkNASTransport
		for _, ie := range down.ProtocolIEs.List {
			if ie.Id.Value == ngapType.ProtocolIEIDAMFUENGAPID {
				h.amfUeId = ie.Value.AMFUENGAPID.Value
				h.logger.Info("updated AMF UE NGAP ID", "id", h.amfUeId)
			}
		}
	case ngapType.ProcedureCodeInitialContextSetup:
		req := ini.Value.InitialContextSetupRequest
		for _, ie := range req.ProtocolIEs.List {
			if ie.Id.Value == ngapType.ProtocolIEIDAMFUENGAPID {
				h.amfUeId = ie.Value.AMFUENGAPID.Value
				h.logger.Info("updated AMF UE NGAP ID", "id", h.amfUeId)
			}
		}
	}
}

func (h *GnbNgapTaskHandler) OnStop(ctx context.Context) error {
	return nil
}
