package nas

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/acore2026/ueransim-go/internal/nas"
	"github.com/acore2026/ueransim-go/internal/security/kdf"
)

type NasCount struct {
	Overflow uint32 // 24-bit
	SQN      uint8  // 8-bit
}

func (c NasCount) Uint32() uint32 {
	return (c.Overflow << 8) | uint32(c.SQN)
}

type SecurityContext struct {
	Kamf    []byte
	KnasInt []byte
	KnasEnc []byte
	UlCount NasCount
	DlCount NasCount

	IntegrityAlgorithm byte
	CipheringAlgorithm byte

	LastDlSQNs []uint8
}

func NewSecurityContext(kNasInt, kNasEnc []byte, integrityAlg, cipheringAlg byte) *SecurityContext {
	return &SecurityContext{
		KnasInt:            kNasInt,
		KnasEnc:            kNasEnc,
		IntegrityAlgorithm: integrityAlg,
		CipheringAlgorithm: cipheringAlg,
		UlCount:            NasCount{},
		DlCount:            NasCount{},
		LastDlSQNs:         make([]uint8, 0, 16),
	}
}

func (sc *SecurityContext) EstimatedDlCount(sn uint8) NasCount {
	count := sc.DlCount
	if count.SQN > sn {
		count.Overflow = (count.Overflow + 1) & 0xFFFFFF
	}
	count.SQN = sn
	return count
}

func (sc *SecurityContext) UpdateDlCount(count NasCount) {
	sc.DlCount = count
}

func (sc *SecurityContext) CheckForReplay(sn uint8) bool {
	for _, s := range sc.LastDlSQNs {
		if s == sn {
			return false
		}
	}
	sc.LastDlSQNs = append(sc.LastDlSQNs, sn)
	if len(sc.LastDlSQNs) > 16 {
		sc.LastDlSQNs = sc.LastDlSQNs[1:]
	}
	return true
}

func (sc *SecurityContext) IncrementUlCount() {
	sc.UlCount.SQN++
	if sc.UlCount.SQN == 0 {
		sc.UlCount.Overflow = (sc.UlCount.Overflow + 1) & 0xFFFFFF
	}
}

func (sc *SecurityContext) Protect(data []byte, headerType nas.SecurityHeaderType) ([]byte, error) {
	direction := byte(0) // Uplink
	bearer := byte(1)    // 3GPP access

	count := sc.UlCount.Uint32()
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
	msgForMac[0] = uint8(count & 0xFF)
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
	res[6] = uint8(count & 0xFF)
	copy(res[7:], payload)

	sc.IncrementUlCount()
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

	direction := byte(1) // Downlink
	bearer := byte(1)    // 3GPP access

	mac := data[2:6]
	sn := uint8(data[6])
	payload := data[7:]

	if !sc.CheckForReplay(sn) {
		return nil, headerType, fmt.Errorf("NAS replay protection triggered for SN %d", sn)
	}

	estimated := sc.EstimatedDlCount(sn)
	count := estimated.Uint32()
	if headerType == nas.SecurityHeaderTypeIntegrityProtectedWithNewSecurityContext {
		count = 0
	}

	msgForMac := make([]byte, 1+len(payload))
	msgForMac[0] = sn
	copy(msgForMac[1:], payload)

	integrityKey := sc.KnasInt
	if (headerType == nas.SecurityHeaderTypeIntegrityProtectedWithNewSecurityContext || headerType == nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNewSecurityContext) && len(sc.Kamf) > 0 && len(payload) > 3 {
		integrityAlg := payload[3] & 0x07
		integrityKey = kdf.DeriveKnas(sc.Kamf, 2, integrityAlg)
	}

	expectedMac, err := NIA2(integrityKey, count, bearer, direction, msgForMac)
	if err != nil {
		return nil, headerType, err
	}

	if !bytes.Equal(mac, expectedMac) {
		return nil, headerType, fmt.Errorf(
			"NAS MAC verification failed: headerType=%d sn=%d received=%s expected=%s payload=%s",
			headerType,
			sn,
			hex.EncodeToString(mac),
			hex.EncodeToString(expectedMac),
			hex.EncodeToString(payload),
		)
	}

	if sc.CipheringAlgorithm != 0 && (headerType == nas.SecurityHeaderTypeIntegrityProtectedAndCiphered || headerType == nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNewSecurityContext) {
		var err error
		payload, err = NEA2(sc.KnasEnc, count, bearer, direction, payload)
		if err != nil {
			return nil, headerType, err
		}
	}

	sc.UpdateDlCount(estimated)
	return payload, headerType, nil
}
