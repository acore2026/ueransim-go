package nas

import (
	"bytes"
	"fmt"
	"github.com/acore2026/ueransim-go/internal/nas"
)

type SecurityContext struct {
	KnasInt []byte
	KnasEnc []byte
	UlCount uint32
	DlCount uint32
	
	IntegrityAlgorithm byte
	CipheringAlgorithm byte
}

func NewSecurityContext(kNasInt, kNasEnc []byte, integrityAlg, cipheringAlg byte) *SecurityContext {
	return &SecurityContext{
		KnasInt:            kNasInt,
		KnasEnc:            kNasEnc,
		IntegrityAlgorithm: integrityAlg,
		CipheringAlgorithm: cipheringAlg,
		UlCount:            0,
		DlCount:            0,
	}
}

func (sc *SecurityContext) Protect(data []byte, headerType nas.SecurityHeaderType) ([]byte, error) {
	direction := byte(1) // Uplink
	bearer := byte(0)
	
	count := sc.UlCount
	if headerType == nas.SecurityHeaderTypeIntegrityProtectedWithNewSecurityContext || headerType == nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNewSecurityContext {
		count = 0
	}
	
	payload := data
	if sc.CipheringAlgorithm != 0 && (headerType == nas.SecurityHeaderTypeIntegrityProtectedAndCiphered || headerType == nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNewSecurityContext) {
		var err error
		payload, err = NEA2(sc.KnasEnc, count, bearer, direction, data)
		if err != nil {
			return nil, err
		}
	}
	
	// TS 24.501 NIA2: MAC over [SN(1)][Ciphered NAS PDU(n)]
	msgForMac := make([]byte, 1+len(payload))
	msgForMac[0] = byte(count & 0xFF)
	copy(msgForMac[1:], payload)
	
	mac, err := NIA2(sc.KnasInt, count, bearer, direction, msgForMac)
	if err != nil {
		return nil, err
	}
	
	// Final: [PD(1)][HeaderType(1)][MAC(4)][SN(1)][Ciphered NAS PDU(n)]
	res := make([]byte, 7+len(payload))
	res[0] = nas.PD_5G_MOBILITY_MANAGEMENT
	res[1] = byte(headerType)
	copy(res[2:6], mac)
	res[6] = byte(count & 0xFF)
	copy(res[7:], payload)
	
	sc.UlCount++
	return res, nil
}

func (sc *SecurityContext) Unprotect(data []byte) ([]byte, nas.SecurityHeaderType, error) {
	if len(data) < 7 {
		return data, nas.SecurityHeaderTypePlainNas, nil
	}
	
	headerType := nas.SecurityHeaderType(data[1] & 0x0F)
	if headerType == nas.SecurityHeaderTypePlainNas {
		return data, headerType, nil
	}
	
	direction := byte(0) // Downlink
	bearer := byte(0)
	
	mac := data[2:6]
	sn := uint32(data[6])
	payload := data[7:]
	
	count := (sc.DlCount & 0xFFFFFF00) | sn
	if headerType == nas.SecurityHeaderTypeIntegrityProtectedWithNewSecurityContext {
		count = 0
	}
	
	msgForMac := make([]byte, 1+len(payload))
	msgForMac[0] = byte(sn)
	copy(msgForMac[1:], payload)
	
	expectedMac, err := NIA2(sc.KnasInt, count, bearer, direction, msgForMac)
	if err != nil {
		return nil, headerType, err
	}
	
	if !bytes.Equal(mac, expectedMac) {
		return nil, headerType, fmt.Errorf("NAS MAC verification failed")
	}
	
	sc.DlCount = count + 1
	return payload, headerType, nil
}
