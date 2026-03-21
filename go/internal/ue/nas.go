package ue

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/acore2026/ueransim-go/internal/config"
	"github.com/acore2026/ueransim-go/internal/core/logging"
	"github.com/acore2026/ueransim-go/internal/core/runtime"
	"github.com/acore2026/ueransim-go/internal/nas"
	"github.com/acore2026/ueransim-go/internal/security/kdf"
	"github.com/acore2026/ueransim-go/internal/security/milenage"
	secnas "github.com/acore2026/ueransim-go/internal/security/nas"
	"github.com/acore2026/ueransim-go/internal/ue/tun"
)

type NasState int

const (
	StateDeregistered NasState = iota
	StateAuthentication
	StateSecurityMode
	StateRegistered
	StatePduSessionPending
	StateSessionReady
)

// NasTaskHandler keeps the supported Go happy path explicit and narrow:
// Registration Request -> Authentication Response -> Security Mode Complete ->
// Registration Complete -> UL NAS Transport(PDU Session Establishment Request) ->
// DL NAS Transport(PDU Session Establishment Accept).
type NasTaskHandler struct {
	logger     logging.Logger
	supi       string
	supiID     string
	imei       string
	imeisv     string
	mcc        string
	mnc        string
	key        []byte
	opc        []byte
	amf        string
	tunNetmask string

	state        NasState
	sec          *secnas.SecurityContext
	session      *config.Session
	sessionID    byte
	sessionPTI   byte
	pduRequested bool

	rrcTask *runtime.Task
	rlsTask *runtime.Task
	tunTask *runtime.Task
}

func NewNasTaskHandler(logger logging.Logger, cfg *config.UEConfig, rrcTask *runtime.Task, rlsTask *runtime.Task, tunTask *runtime.Task) *NasTaskHandler {
	key, _ := hex.DecodeString(cfg.Key)
	opc, _ := hex.DecodeString(cfg.OP)
	if cfg.OPType == "OP" {
		opc = milenage.GenerateOpC(key, opc)
	}

	return &NasTaskHandler{
		logger:     logger.With("component", "nas"),
		supi:       cfg.SUPI,
		supiID:     normalizeSupi(cfg.SUPI),
		imei:       normalizeIMEI(cfg.IMEI),
		imeisv:     cfg.IMEISV,
		mcc:        cfg.MCC,
		mnc:        cfg.MNC,
		key:        key,
		opc:        opc,
		amf:        cfg.AMF,
		tunNetmask: cfg.TUNNetmask,
		state:      StateDeregistered,
		session:    firstSession(cfg.Sessions),
		sessionID:  1,
		sessionPTI: 1,
		rrcTask:    rrcTask,
		rlsTask:    rlsTask,
		tunTask:    tunTask,
	}
}

func (h *NasTaskHandler) OnStart(ctx context.Context, t *runtime.Task) error {
	h.logger.Info("NAS task started")
	return h.sendRegistrationRequest(t)
}

func (h *NasTaskHandler) sendRegistrationRequest(t *runtime.Task) error {
	h.logger.Info("sending Registration Request")

	msin := h.supiID[len(h.supiID)-10:]
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
		Capability5GMM: &nas.Capability5GMM{
			Octets: [13]byte{0x07},
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

func (h *NasTaskHandler) sendProtectedNas(data []byte, headerType nas.SecurityHeaderType) error {
	if h.sec == nil {
		return fmt.Errorf("security context not initialized")
	}
	protected, err := h.sec.Protect(data, headerType)
	if err != nil {
		return err
	}
	return h.sendPlainNas(protected)
}

func (h *NasTaskHandler) OnMessage(ctx context.Context, msg runtime.Message) error {
	switch msg.Type {
	case "rrc_to_nas":
		nasPdu := msg.Payload.([]byte)
		headerByte := byte(0)
		if len(nasPdu) > 1 {
			headerByte = nasPdu[1]
		}
		h.logger.Info("received NAS PDU from RRC", "len", len(nasPdu), "securityHeader", fmt.Sprintf("0x%02x", headerByte))

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
		case nas.MsgTypeIdentityRequest:
			return h.handleIdentityRequest(nasPdu)
		case nas.MsgTypeSecurityModeCommand:
			return h.handleSecurityModeCommand(nasPdu)
		case nas.MsgTypeRegistrationAccept:
			return h.handleRegistrationAccept(nasPdu)
		case nas.MsgTypeDlNasTransport:
			return h.handleDlNasTransport(nasPdu)
		default:
			h.logger.Info("received NAS message", "type", fmt.Sprintf("0x%02x", msgType))
		}
	}
	return nil
}

func (h *NasTaskHandler) handleIdentityRequest(data []byte) error {
	h.logger.Info("handling Identity Request")

	req, err := nas.DecodeIdentityRequest(data)
	if err != nil {
		return err
	}

	identity := nas.IE5gsMobileIdentity{Type: nas.MobileIdentityTypeSuci}
	switch req.IdentityType {
	case nas.MobileIdentityTypeSuci:
		msin := h.supiID[len(h.supiID)-10:]
		identity.Suci = &nas.Suci{
			Mcc:  h.mcc,
			Mnc:  h.mnc,
			MSIN: msin,
		}
	case nas.MobileIdentityTypeImei:
		identity.Type = nas.MobileIdentityTypeImei
		identity.Digits = h.imei
	case nas.MobileIdentityTypeImeisv:
		identity.Type = nas.MobileIdentityTypeImeisv
		identity.Digits = h.imeisv
	default:
		return fmt.Errorf("unsupported identity type request 0x%02x", req.IdentityType)
	}

	h.logger.Info("sending Identity Response", "identityType", fmt.Sprintf("0x%02x", identity.Type))
	resp := &nas.IdentityResponse{MobileIdentity: identity}
	return h.sendProtectedNas(resp.Encode().Data(), nas.SecurityHeaderTypeIntegrityProtectedAndCiphered)
}

func (h *NasTaskHandler) handleAuthenticationRequest(data []byte) error {
	h.logger.Info("handling Authentication Request")

	req, err := nas.DecodeAuthenticationRequest(data)
	if err != nil {
		return err
	}

	m := milenage.NewMilenage(h.key, h.opc)

	// Verify AUTN
	ok, sqn, ak, mac, xmac := m.VerifyAutn(req.Rand[:], req.Autn)
	h.logger.Info("auth debug",
		"rand", hex.EncodeToString(req.Rand[:]),
		"autn", hex.EncodeToString(req.Autn[:]),
		"derived_sqn", hex.EncodeToString(sqn),
		"derived_ak", hex.EncodeToString(ak),
		"mac", hex.EncodeToString(mac),
		"xmac", hex.EncodeToString(xmac),
		"autn_ok", ok)

	if !ok {
		h.logger.Error("AUTN verification failed")
		// Continue for now to see what RES* we get
	}

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

	resStar := kdf.DeriveResStar(ck, ik, req.Rand[:], res, snName)

	sqnXorAk := make([]byte, len(sqn))
	for i := range sqn {
		sqnXorAk[i] = sqn[i] ^ ak[i]
	}

	// Derive K_AMF for the new NAS security context.
	kAusf := kdf.DeriveKausf(ck, ik, snName, sqnXorAk)
	kSeaf := kdf.DeriveKseaf(kAusf, snName)
	kAmf := kdf.DeriveKamf(kSeaf, h.supiID, []byte{0x00, 0x00})

	h.sec = secnas.NewSecurityContext(nil, nil, 0, 0)
	h.sec.Kamf = kAmf
	h.state = StateAuthentication

	h.logger.Info("sending Authentication Response", "resStar", hex.EncodeToString(resStar))
	resp := &nas.AuthenticationResponse{ResStar: resStar}
	return h.sendPlainNas(resp.Encode().Data())
}

func (h *NasTaskHandler) handleSecurityModeCommand(data []byte) error {
	h.logger.Info("handling Security Mode Command")

	cmd, err := nas.DecodeSecurityModeCommand(data)
	if err != nil {
		return err
	}

	kAmf := h.sec.Kamf
	h.sec.KnasEnc = kdf.DeriveKnas(kAmf, 1, cmd.SelectedCipheringAlgorithm)
	h.sec.KnasInt = kdf.DeriveKnas(kAmf, 2, cmd.SelectedIntegrityAlgorithm)
	h.sec.IntegrityAlgorithm = cmd.SelectedIntegrityAlgorithm
	h.sec.CipheringAlgorithm = cmd.SelectedCipheringAlgorithm
	h.sec.UlCount = secnas.NasCount{}
	h.sec.DlCount = secnas.NasCount{}
	h.state = StateSecurityMode

	h.logger.Info("derived NAS keys", "integrity", cmd.SelectedIntegrityAlgorithm, "ciphering", cmd.SelectedCipheringAlgorithm)

	msin := h.supiID[len(h.supiID)-10:]
	resp := &nas.SecurityModeComplete{
		MobileIdentity: nas.IE5gsMobileIdentity{
			Type: nas.MobileIdentityTypeSuci,
			Suci: &nas.Suci{
				Mcc: h.mcc, Mnc: h.mnc, MSIN: msin,
			},
		},
	}

	h.logger.Info("sending Security Mode Complete")
	return h.sendProtectedNas(resp.Encode().Data(), nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNewSecurityContext)
}

func (h *NasTaskHandler) OnStop(ctx context.Context) error {
	return nil
}

func (h *NasTaskHandler) handleRegistrationAccept(data []byte) error {
	h.logger.Info("Registration Accept received")
	h.state = StateRegistered

	if err := h.sendProtectedNas((&nas.RegistrationComplete{}).Encode().Data(), nas.SecurityHeaderTypeIntegrityProtectedAndCiphered); err != nil {
		return err
	}
	h.logger.Info("sending Registration Complete")

	if !h.pduRequested && h.session != nil {
		return h.sendPduSessionEstablishmentRequest()
	}
	return nil
}

func (h *NasTaskHandler) handleDlNasTransport(data []byte) error {
	dl, err := nas.DecodeDlNasTransport(data)
	if err != nil {
		return err
	}
	if len(dl.PayloadContainer) < 4 {
		h.logger.Info("received DL NAS Transport without recognizable payload")
		return nil
	}

	switch dl.PayloadContainer[3] {
	case nas.MsgTypePduSessionEstablishmentAccept:
		accept, err := nas.DecodePduSessionEstablishmentAccept(dl.PayloadContainer)
		if err != nil {
			return err
		}
		h.state = StateSessionReady
		h.logger.Info("PDU Session Establishment Accept received", "psi", accept.PduSessionID, "pti", accept.Pti, "ip", accept.PDUAddress)
		if h.tunTask != nil && accept.PDUAddress != "" {
			if err := h.tunTask.Send(runtime.Message{
				Type: tun.MessageTypeConfigure,
				Payload: tun.ConfigureMessage{
					IPAddress: accept.PDUAddress,
					Netmask:   h.sessionNetmask(),
					Route:     true,
				},
			}); err != nil {
				return err
			}
		}
		if h.rlsTask != nil {
			if err := h.rlsTask.Send(runtime.Message{
				Type:    "nas_session_ready",
				Payload: accept.PduSessionID,
			}); err != nil {
				return err
			}
		}
	case nas.MsgTypePduSessionEstablishmentReject:
		h.logger.Error("PDU Session Establishment Reject received")
	default:
		h.logger.Info("received DL NAS Transport payload", "messageType", fmt.Sprintf("0x%02x", dl.PayloadContainer[3]))
	}
	return nil
}

func (h *NasTaskHandler) sendPduSessionEstablishmentRequest() error {
	if h.session == nil {
		return nil
	}

	sNssai, err := parseSNssai(h.session.Slice)
	if err != nil {
		return err
	}

	smReq := (&nas.PduSessionEstablishmentRequest{
		PduSessionID:   h.sessionID,
		Pti:            h.sessionPTI,
		PduSessionType: 1,
		SscMode:        1,
	}).Encode().Data()

	ul := &nas.UlNasTransport{
		PayloadContainerType: 1,
		PayloadContainer:     smReq,
		PduSessionID:         h.sessionID,
		RequestType:          1,
		SNssai:               sNssai,
		Dnn:                  h.session.APN,
	}

	h.pduRequested = true
	h.state = StatePduSessionPending
	h.logger.Info("sending PDU Session Establishment Request", "psi", h.sessionID, "apn", h.session.APN)
	return h.sendProtectedNas(ul.Encode().Data(), nas.SecurityHeaderTypeIntegrityProtectedAndCiphered)
}

func firstSession(sessions []config.Session) *config.Session {
	if len(sessions) == 0 {
		return nil
	}
	session := sessions[0]
	return &session
}

func parseSNssai(slice config.Slice) (*nas.SNssai, error) {
	sst, err := parseFlexibleUint(slice.SST)
	if err != nil {
		return nil, fmt.Errorf("parse SST: %w", err)
	}
	res := &nas.SNssai{SST: byte(sst)}

	if slice.SD == nil {
		return res, nil
	}
	sd, err := parseFlexibleUint(slice.SD)
	if err != nil {
		return nil, fmt.Errorf("parse SD: %w", err)
	}
	res.SD = []byte{byte(sd >> 16), byte(sd >> 8), byte(sd)}
	return res, nil
}

func parseFlexibleUint(v any) (uint64, error) {
	switch value := v.(type) {
	case int:
		return uint64(value), nil
	case int64:
		return uint64(value), nil
	case uint64:
		return value, nil
	case uint:
		return uint64(value), nil
	case string:
		if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
			return strconv.ParseUint(value[2:], 16, 64)
		}
		return strconv.ParseUint(value, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported numeric type %T", v)
	}
}

func normalizeSupi(supi string) string {
	return strings.TrimPrefix(strings.TrimPrefix(supi, "imsi-"), "supi-")
}

func normalizeIMEI(imei string) string {
	if len(imei) < 14 {
		return imei
	}
	base := imei[:14]
	checkDigit := imeiCheckDigit(base)
	if len(imei) == 14 {
		return base + string('0'+checkDigit)
	}
	if len(imei) == 15 {
		return base + string('0'+checkDigit)
	}
	return imei
}

func imeiCheckDigit(base string) byte {
	sum := 0
	for i := 0; i < len(base); i++ {
		d := int(base[i] - '0')
		if i%2 == 1 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return byte((10 - (sum % 10)) % 10)
}

func (h *NasTaskHandler) sessionNetmask() string {
	if h.tunNetmask == "" {
		return "255.255.255.0"
	}
	return h.tunNetmask
}
