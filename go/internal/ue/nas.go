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
	secnas "github.com/acore2026/ueransim-go/internal/security/nas"
)

type NasState int

const (
	StateDeregistered NasState = iota
	StateAuthentication
	StateSecurityMode
	StateRegistered
)

type NasTaskHandler struct {
	logger logging.Logger
	supi   string
	mcc    string
	mnc    string
	key    []byte
	opc    []byte
	
	state NasState
	sec   *secnas.SecurityContext
	
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
		state:   StateDeregistered,
		rrcTask: rrcTask,
	}
}

func (h *NasTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.logger.Info("NAS task started")
	return h.sendRegistrationRequest(t)
}

func (h *NasTaskHandler) sendRegistrationRequest(t *runtime.Task) error {
	h.logger.Info("sending Registration Request")
	
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
	
	return h.sendPlainNas(req.Encode().Data())
}

func (h *NasTaskHandler) sendPlainNas(data []byte) error {
	return h.rrcTask.Send(runtime.Message{
		Type:    "nas_to_rrc",
		Payload: data,
	})
}

func (h *NasTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case "rrc_to_nas":
		nasPdu := msg.Payload.([]byte)
		h.logger.Info("received NAS PDU from RRC", "hex", hex.EncodeToString(nasPdu))
		
		// Unprotect if needed
		if h.sec != nil && len(nasPdu) > 7 && nasPdu[1] != 0 {
			var err error
			nasPdu, _, err = h.sec.Unprotect(nasPdu)
			if err != nil {
				return err
			}
		}
		
		if len(nasPdu) < 3 {
			return nil
		}
		
		msgType := nasPdu[2]
		switch msgType {
		case nas.MsgTypeAuthenticationRequest:
			return h.handleAuthenticationRequest(nasPdu)
		case nas.MsgTypeSecurityModeCommand:
			return h.handleSecurityModeCommand(nasPdu)
		case nas.MsgTypeRegistrationAccept:
			h.logger.Info("Registration Accept received! SUCCESS")
			h.state = StateRegistered
		default:
			h.logger.Info("received NAS message", "type", fmt.Sprintf("0x%02x", msgType))
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
	
	m := milenage.NewMilenage(h.key, h.opc)
	res, ck, ik, _, _ := m.F2345(req.Rand[:])
	
	mcc := h.mcc
	for len(mcc) < 3 {
		mcc = "0" + mcc
	}
	mnc := h.mnc
	for len(mnc) < 3 {
		mnc = "0" + mnc
	}
	snName := fmt.Sprintf("5G:mnc%s.mcc%s.3gppnetwork.org", mnc, mcc)
	
	h.logger.Info("auth debug", 
		"rand", hex.EncodeToString(req.Rand[:]),
		"autn", hex.EncodeToString(req.Autn[:]),
		"ck", hex.EncodeToString(ck),
		"ik", hex.EncodeToString(ik),
		"res", hex.EncodeToString(res),
		"snName", snName)
	
	resStar := kdf.DeriveResStar(ck, ik, req.Rand[:], res, snName)
	
	// Derive K_AMF for future security context
	kSeaf := kdf.DeriveKseaf(ck, ik, snName)
	kAmf := kdf.DeriveKamf(kSeaf, h.supi, []byte{0x00, 0x00}) 
	
	// Store K_AMF temporarily in a new security context
	h.sec = secnas.NewSecurityContext(nil, nil, 0, 0)
	h.sec.KnasInt = kAmf 
	
	h.logger.Info("sending Authentication Response")
	resp := &nas.AuthenticationResponse{ResStar: resStar}
	return h.sendPlainNas(resp.Encode().Data())
}

func (h *NasTaskHandler) handleSecurityModeCommand(data []byte) error {
	h.logger.Info("handling Security Mode Command")
	
	cmd, err := nas.DecodeSecurityModeCommand(data)
	if err != nil {
		return err
	}
	
	// Derive NAS keys from K_AMF (currently stored in h.sec.KnasInt)
	kAmf := h.sec.KnasInt
	h.sec.KnasEnc = kdf.DeriveKnas(kAmf, 1, cmd.SelectedCipheringAlgorithm)
	h.sec.KnasInt = kdf.DeriveKnas(kAmf, 2, cmd.SelectedIntegrityAlgorithm)
	h.sec.IntegrityAlgorithm = cmd.SelectedIntegrityAlgorithm
	h.sec.CipheringAlgorithm = cmd.SelectedCipheringAlgorithm
	
	h.logger.Info("derived NAS keys", "integrity", cmd.SelectedIntegrityAlgorithm, "ciphering", cmd.SelectedCipheringAlgorithm)
	
	msin := h.supi[len(h.supi)-10:]
	resp := &nas.SecurityModeComplete{
		MobileIdentity: nas.IE5gsMobileIdentity{
			Type: nas.MobileIdentityTypeSuci,
			Suci: &nas.Suci{
				Mcc: h.mcc, Mnc: h.mnc, MSIN: msin,
			},
		},
	}
	
	encoded := resp.Encode().Data()
	protected, err := h.sec.Protect(encoded, nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNewSecurityContext)
	if err != nil {
		return err
	}
	
	h.logger.Info("sending Security Mode Complete")
	return h.rrcTask.Send(runtime.Message{
		Type:    "nas_to_rrc",
		Payload: protected,
	})
}

func (h *NasTaskHandler) OnStop(ctx context.Context) error {
	return nil
}
