package nas

import (
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
	// 1. Ciphering (if required)
	// For simplicity, we implement plain integrity first as many SMC responses are just integrity protected
	
	// 2. Integrity
	if sc.IntegrityAlgorithm != 0 {
		// NIA2 (AES-CMAC)
		mac, err := NIA2(sc.KnasInt, sc.UlCount, 0, 0, data)
		if err != nil {
			return nil, err
		}
		
		// Build Protected NAS PDU
		// [PD(1)][Security Header(1)][MAC(4)][Sequence Number(1)][NAS PDU(n)]
		res := make([]byte, 7+len(data))
		res[0] = nas.PD_5G_MOBILITY_MANAGEMENT
		res[1] = byte(headerType)
		copy(res[2:6], mac)
		res[6] = byte(sc.UlCount & 0xFF)
		copy(res[7:], data)
		
		sc.UlCount++
		return res, nil
	}
	
	return data, nil
}

func (sc *SecurityContext) Unprotect(data []byte) ([]byte, nas.SecurityHeaderType, error) {
	if len(data) < 7 {
		return data, nas.SecurityHeaderTypePlainNas, nil
	}
	
	headerType := nas.SecurityHeaderType(data[1] & 0x0F)
	if headerType == nas.SecurityHeaderTypePlainNas {
		return data, headerType, nil
	}
	
	// For now, we just strip the security header and return the inner NAS PDU
	// In a real implementation, we would verify MAC and decipher
	
	innerNas := data[7:]
	sc.DlCount++
	return innerNas, headerType, nil
}
