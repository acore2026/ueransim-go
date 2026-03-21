package ue

import (
	"context"
	"fmt"
	"time"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/rlc"
	"github.com/acore2026/ueransim-go/internal/rrc"
)

type RrcState int

const (
	StateIdle RrcState = iota
	StateConnecting
	StateConnected
)

type RrcTaskHandler struct {
	logger      logging.Logger
	rlcTask     *runtime.Task
	nasTask     *runtime.Task
	isFirstResp bool

	state     RrcState
	nasBuffer [][]byte
	t300Timer *time.Timer
	thisTask  *runtime.Task
}

func NewRrcTaskHandler(logger logging.Logger, rlcTask *runtime.Task, nasTask *runtime.Task) *RrcTaskHandler {
	return &RrcTaskHandler{
		logger:      logger.With("component", "rrc"),
		rlcTask:     rlcTask,
		nasTask:     nasTask,
		isFirstResp: true,
		state:       StateIdle,
		nasBuffer:   make([][]byte, 0),
	}
}

func (h *RrcTaskHandler) startT300() {
	h.stopT300()
	h.t300Timer = time.AfterFunc(10*time.Second, func() {
		if h.thisTask != nil {
			_ = h.thisTask.Send(runtime.Message{Type: "t300_expiry"})
		}
	})
}

func (h *RrcTaskHandler) stopT300() {
	if h.t300Timer != nil {
		h.t300Timer.Stop()
		h.t300Timer = nil
	}
}

func (h *RrcTaskHandler) SetNasTask(t *runtime.Task) {
	h.nasTask = t
}

func (h *RrcTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.logger.Info("RRC task started")
	h.thisTask = t
	return nil
}

func (h *RrcTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case "nas_to_rrc":
		return h.handleNasToRrc(ctx, msg.Payload.([]byte))
	case "rlc_to_rrc":
		return h.handleRlsToRrc(ctx, msg.Payload.([]byte))
	case "t300_expiry":
		h.logger.Warn("T300 timer expired, connection failed")
		h.state = StateIdle
		h.stopT300()
		return nil
	}
	return nil
}

func (h *RrcTaskHandler) handleNasToRrc(ctx context.Context, nasPdu []byte) error {
	h.logger.Info("received NAS PDU from NAS layer", "len", len(nasPdu), "state", h.state)

	switch h.state {
	case StateIdle:
		h.logger.Info("triggering RRC connection establishment")
		h.nasBuffer = append(h.nasBuffer, nasPdu)
		h.state = StateConnecting

		rrcPdu := rrc.BuildRRCSetupRequest(0x123456789A)
		err := h.rlcTask.Send(runtime.Message{
			Type: "upper_to_rlc",
			Payload: rlc.UpperToRlcMessage{
				Mode: rlc.ModeTM,
				Pdu:  rrcPdu,
			},
		})
		if err != nil {
			return err
		}

		h.startT300()
		return nil

	case StateConnecting:
		h.logger.Info("buffering NAS PDU while connecting")
		h.nasBuffer = append(h.nasBuffer, nasPdu)
		return nil

	case StateConnected:
		h.logger.Info("sending NAS PDU over established connection")
		rrcPdu := rrc.BuildULInformationTransfer(nasPdu)
		return h.rlcTask.Send(runtime.Message{
			Type: "upper_to_rlc",
			Payload: rlc.UpperToRlcMessage{
				Mode: rlc.ModeUM,
				Pdu:  rrcPdu,
			},
		})
	}
	return nil
}

func (h *RrcTaskHandler) handleRlsToRrc(ctx context.Context, rrcPdu []byte) error {
	if len(rrcPdu) == 0 {
		return nil
	}

	h.logger.Info("received RRC PDU from RLC", "hex", fmt.Sprintf("%02x", rrcPdu[0]), "state", h.state)

	if rrcPdu[0] == 0x20 {
		h.logger.Info("detected RRCSetup message")
		if h.state == StateConnecting {
			h.state = StateConnected
			h.stopT300()
			h.logger.Info("RRC connection established")

			if len(h.nasBuffer) > 0 {
				firstNas := h.nasBuffer[0]
				h.nasBuffer = h.nasBuffer[1:]

				h.logger.Info("sending RRCSetupComplete with first buffered NAS PDU")
				resp := rrc.BuildRRCSetupComplete(firstNas)
				if err := h.rlcTask.Send(runtime.Message{
					Type: "upper_to_rlc",
					Payload: rlc.UpperToRlcMessage{
						Mode: rlc.ModeUM,
						Pdu:  resp,
					},
				}); err != nil {
					return err
				}

				for _, nas := range h.nasBuffer {
					h.logger.Info("sending subsequent buffered NAS PDU in ULInformationTransfer")
					resp := rrc.BuildULInformationTransfer(nas)
					if err := h.rlcTask.Send(runtime.Message{
						Type: "upper_to_rlc",
						Payload: rlc.UpperToRlcMessage{
							Mode: rlc.ModeUM,
							Pdu:  resp,
						},
					}); err != nil {
						return err
					}
				}
				h.nasBuffer = make([][]byte, 0)
			}
		}
		return nil
	}

	if len(rrcPdu) > 0 {
		index := (rrcPdu[0] >> 3) & 0x0F
		switch index {
		case 0:
			h.logger.Info("detected RRCReconfiguration message")
			resp := rrc.BuildRRCReconfigurationComplete()
			return h.rlcTask.Send(runtime.Message{
				Type: "upper_to_rlc",
				Payload: rlc.UpperToRlcMessage{
					Mode: rlc.ModeUM,
					Pdu:  resp,
				},
			})

		case 5:
			h.logger.Info("detected DLInformationTransfer message")
			if len(rrcPdu) > 2 {
				nasLen := int(rrcPdu[1])
				if len(rrcPdu) >= 2+nasLen {
					nasPdu := rrcPdu[2 : 2+nasLen]
					h.logger.Info("forwarding NAS PDU to NAS task")
					return h.nasTask.Send(runtime.Message{Type: "rrc_to_nas", Payload: nasPdu})
				}
			}

		case 2:
			h.logger.Info("detected RRCRelease message")
			h.state = StateIdle
			h.stopT300()
			return nil
		}
	}

	return nil
}

func (h *RrcTaskHandler) OnStop(context.Context) error {
	h.stopT300()
	return nil
}
