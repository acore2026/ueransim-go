package ue

import (
	"context"
	"encoding/hex"

	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/nas"
)

type NasTaskHandler struct {
	logger logging.Logger
	supi   string
	mcc    string
	mnc    string
	
	rrcTask *runtime.Task
}

func NewNasTaskHandler(logger logging.Logger, supi, mcc, mnc string, rrcTask *runtime.Task) *NasTaskHandler {
	return &NasTaskHandler{
		logger:  logger.With("component", "nas"),
		supi:    supi,
		mcc:     mcc,
		mnc:     mnc,
		rrcTask: rrcTask,
	}
}

func (h *NasTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.logger.Info("NAS task started")
	
	// Initial action: Trigger Registration
	return h.sendRegistrationRequest(t)
}

func (h *NasTaskHandler) sendRegistrationRequest(t *runtime.Task) error {
	h.logger.Info("sending Registration Request")
	
	// Construct the NAS Registration Request
	// Simplified MSIN extraction from SUPI (e.g., "imsi-208930123456789")
	msin := h.supi[len(h.supi)-10:]
	
	req := &nas.RegistrationRequest{
		RegistrationType: nas.IE5gsRegistrationType{
			FollowOnRequest:  true,
			RegistrationType: 0x01,
		},
		NasKeySetIdentifier: nas.IENasKeySetIdentifier{
			KeySetIdentifier: 0x07,
		},
		MobileIdentity: nas.IE5gsMobileIdentity{
			Type: nas.MobileIdentityTypeSuci,
			Suci: &nas.Suci{
				Mcc:  h.mcc,
				Mnc:  h.mnc,
				MSIN: msin,
			},
		},
		UeSecurityCapability: &nas.UeSecurityCapability{
			EA0: true, EA1: true, EA2: true,
			IA0: true, IA1: true, IA2: true,
		},
	}
	
	buf := req.Encode()
	
	// Send to RRC task for delivery over radio
	return h.rrcTask.Send(runtime.Message{
		Type:    "nas_to_rrc",
		Payload: buf.Data(),
	})
}

func (h *NasTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case "rrc_to_nas":
		h.logger.Info("received NAS PDU from RRC")
		nasPdu := msg.Payload.([]byte)
		// For now just log it
		h.logger.Info("received NAS message", "len", len(nasPdu), "data", hex.EncodeToString(nasPdu))
	}
	return nil
}

func (h *NasTaskHandler) OnStop(ctx context.Context) error {
	return nil
}
