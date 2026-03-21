package rlc

import (
	"context"
	"encoding/hex"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/gnbctx"
	"github.com/acore2026/ueransim-go/internal/rls"
	"github.com/acore2026/ueransim-go/internal/rrc"
)

type RlcTaskHandler struct {
	logger   logging.Logger
	rlsTask  *runtime.Task
	rrcTask  *runtime.Task // For DCCH (UM) - UE: RRC, gNB: NGAP
	nasTask  *runtime.Task // For DTCH (UM) - UE: TUN, gNB: GTP
	ccchTask *runtime.Task // For CCCH (TM) - UE: RRC, gNB: RLS
	txSN     uint8
	rxSN     uint8
}

type UpperToRlcMessage struct {
	Mode      Mode
	Pdu       []byte
	SessionID uint8 // For Data type
}

type RlsToRlcMessage struct {
	PduType rls.PduType
	Pdu     []byte
	Payload uint32 // From RLS header (LCID or PSI)
}

type RlcToRlsMessage struct {
	PduType rls.PduType
	Pdu     []byte
	Payload uint32 // LCID for RRC, PSI for Data
}

const (
	LcidCCCH uint32 = 0
	LcidDCCH uint32 = 1
)

func NewRlcTaskHandler(logger logging.Logger, rlsTask *runtime.Task) *RlcTaskHandler {
	return &RlcTaskHandler{
		logger:  logger.With("component", "rlc"),
		rlsTask: rlsTask,
	}
}

func (h *RlcTaskHandler) SetRrcTask(t *runtime.Task) {
	h.rrcTask = t
}

func (h *RlcTaskHandler) SetNasTask(t *runtime.Task) {
	h.nasTask = t
}

func (h *RlcTaskHandler) SetCcchTask(t *runtime.Task) {
	h.ccchTask = t
}

func (h *RlcTaskHandler) SetTunTask(t *runtime.Task) {
	h.nasTask = t
}

func (h *RlcTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.logger.Info("RLC task started")
	return nil
}

func (h *RlcTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case "upper_to_rlc":
		payload := msg.Payload.(UpperToRlcMessage)
		return h.handleUpperToRlc(ctx, payload)
	case "rls_to_rlc":
		payload := msg.Payload.(RlsToRlcMessage)
		return h.handleRlsToRlc(ctx, payload)
	}
	return nil
}

func (h *RlcTaskHandler) handleUpperToRlc(ctx context.Context, msg UpperToRlcMessage) error {
	var encoded []byte
	pduType := rls.PDU_TYPE_RRC
	payload := LcidCCCH

	if msg.Mode == ModeTM {
		h.logger.Info("encapsulating PDU in RLC-TM", "len", len(msg.Pdu))
		encoded = msg.Pdu
		payload = LcidCCCH
	} else if msg.Mode == ModeUM {
		h.logger.Info("encapsulating PDU in RLC-UM", "len", len(msg.Pdu), "sn", h.txSN)
		header := UMHeader{SI: SIComplete, SN: h.txSN}
		encoded = make([]byte, 1+len(msg.Pdu))
		encoded[0] = header.Encode()
		copy(encoded[1:], msg.Pdu)
		h.txSN = (h.txSN + 1) & 0x3F

		if len(msg.Pdu) > 20 && (msg.Pdu[0] == 0x45 || msg.Pdu[0] == 0x60) {
			pduType = rls.PduTypeData
			payload = uint32(msg.SessionID)
		} else {
			payload = LcidDCCH
		}
	}

	return h.rlsTask.Send(runtime.Message{
		Type: "rlc_to_rls",
		Payload: RlcToRlsMessage{
			PduType: pduType,
			Pdu:     encoded,
			Payload: payload,
		},
	})
}

func (h *RlcTaskHandler) handleRlsToRlc(ctx context.Context, msg RlsToRlcMessage) error {
	pdu := msg.Pdu
	if len(pdu) == 0 {
		return nil
	}

	if msg.PduType == rls.PDU_TYPE_RRC {
		if msg.Payload == LcidCCCH {
			h.logger.Info("decapsulating PDU in RLC-TM (CCCH)", "len", len(pdu))
			if h.ccchTask != nil {
				return h.ccchTask.Send(runtime.Message{
					Type:    "rlc_to_ccch",
					Payload: pdu,
				})
			}
			if h.rrcTask != nil {
				return h.rrcTask.Send(runtime.Message{
					Type:    "rlc_to_rrc",
					Payload: pdu,
				})
			}
		} else {
			// Assume UM (DCCH)
			h.logger.Info("decapsulating PDU in RLC-UM (DCCH)", "len", len(pdu))
			sdu := pdu[1:]
			if h.rrcTask != nil {
				typeStr := "rlc_to_rrc"
				payload := sdu
				if h.ccchTask != nil { // gNB
					typeStr = "rls_to_ngap"
					// Extract NAS from RRC container
					nasPdu := rrc.ExtractNasPdu(sdu)
					if nasPdu != nil {
						payload = nasPdu
						h.logger.Info("extracted NAS PDU from RRC", "hex", hex.EncodeToString(nasPdu))
					} else {
						h.logger.Warn("failed to extract NAS PDU from RRC container", "hex", hex.EncodeToString(sdu))
					}
				}
				return h.rrcTask.Send(runtime.Message{
					Type:    runtime.MessageType(typeStr),
					Payload: payload,
				})
			}
		}
	} else if msg.PduType == rls.PduTypeData {
		h.logger.Info("decapsulating PDU in RLC-UM (Data)", "len", len(pdu))
		sdu := pdu[1:]
		if h.nasTask != nil {
			typeStr := "app_to_tun"
			var payload any = sdu
			if h.ccchTask != nil { // gNB
				typeStr = "rls_to_gtp"
				payload = gnbctx.UplinkPacket{
					SessionID: uint8(msg.Payload),
					Data:      sdu,
				}
			}
			return h.nasTask.Send(runtime.Message{
				Type:    runtime.MessageType(typeStr),
				Payload: payload,
			})
		}
	}

	return nil
}

func (h *RlcTaskHandler) OnStop(ctx context.Context) error {
	return nil
}
