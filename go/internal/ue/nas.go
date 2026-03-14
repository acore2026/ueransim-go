package ue

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/acore2026/ueransim-go/internal/config"
	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/nas"
	"github.com/acore2026/ueransim-go/internal/security/kdf"
	"github.com/acore2026/ueransim-go/internal/security/milenage"
)

type NasTaskHandler struct {
	logger logging.Logger
	supi   string
	mcc    string
	mnc    string
	key    []byte
	opc    []byte
	
	rrcTask *runtime.Task
}

func NewNasTaskHandler(logger logging.Logger, cfg *config.UEConfig, rrcTask *runtime.Task) *NasTaskHandler {
	key, _ := hex.DecodeString(cfg.Key)
	opc, _ := hex.DecodeString(cfg.OP)
	if cfg.OPType == "OP" {
		opc = milenage.GenerateOpC(key, opc)
	}

	return &NasTaskHandler{
		logger:  logger.With("component", "nas"),
		supi:    cfg.SUPI,
		mcc:     cfg.MCC,
		mnc:     cfg.MNC,
		key:     key,
		opc:     opc,
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
		
		if len(nasPdu) < 3 {
			return nil
		}
		
		msgType := nasPdu[2]
		switch msgType {
		case nas.MsgTypeAuthenticationRequest:
			return h.handleAuthenticationRequest(nasPdu)
		default:
			h.logger.Info("received NAS message", "type", fmt.Sprintf("0x%02x", msgType), "len", len(nasPdu))
		}
	}
	return nil
}

func (h *NasTaskHandler) handleAuthenticationRequest(data []byte) error {
	h.logger.Info("handling Authentication Request")
	
	req, err := nas.DecodeAuthenticationRequest(data)
	if err != nil {
		return err
	}
	
	// Run Milenage
	m := milenage.NewMilenage(h.key, h.opc)
	res, ck, ik, _, _ := m.F2345(req.Rand[:])
	
	// Derive RES*
	snName := "5G:mnc093.mcc208.3gppnetwork.org"
	resStar := kdf.DeriveResStar(ck, ik, req.Rand[:], res, snName)
	
	h.logger.Info("sending Authentication Response", "resStar", hex.EncodeToString(resStar))
	
	resp := &nas.AuthenticationResponse{
		ResStar: resStar,
	}
	
	return h.rrcTask.Send(runtime.Message{
		Type:    "nas_to_rrc",
		Payload: resp.Encode().Data(),
	})
}

func (h *NasTaskHandler) OnStop(ctx context.Context) error {
	return nil
}
