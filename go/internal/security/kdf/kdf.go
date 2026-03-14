package kdf

import (
	"crypto/hmac"
	"crypto/sha256"
)

// KDF implements the 3GPP Key Derivation Function (TS 33.220)
// S = FC || P0 || L0 || P1 || L1 || ... || Pn || Ln
func KDF(key []byte, fc byte, p []string, l []int) []byte {
	s := []byte{fc}
	for i := 0; i < len(p); i++ {
		s = append(s, []byte(p[i])...)
		s = append(s, byte(l[i]>>8), byte(l[i]))
	}

	h := hmac.New(sha256.New, key)
	h.Write(s)
	return h.Sum(nil)
}

// KDFBytes is the same as KDF but P are raw bytes
func KDFBytes(key []byte, fc byte, p [][]byte, l []int) []byte {
	s := []byte{fc}
	for i := 0; i < len(p); i++ {
		s = append(s, p[i]...)
		s = append(s, byte(l[i]>>8), byte(l[i]))
	}

	h := hmac.New(sha256.New, key)
	h.Write(s)
	return h.Sum(nil)
}

// DeriveResStar derives RES* from CK, IK (TS 33.501, Annex A.4)
func DeriveResStar(ck, ik []byte, rand, res []byte, snName string) []byte {
	// Key is concatenation of CK and IK
	key := append(ck, ik...)
	
	p := [][]byte{
		[]byte(snName),
		rand,
		res,
	}
	l := []int{
		len(snName),
		len(rand),
		len(res),
	}
	
	// FC = 0x6B
	k := KDFBytes(key, 0x6B, p, l)
	
	// Return the last 16 bytes
	return k[16:]
}
func DeriveKamf(kSeaf []byte, supi string, abba []byte) []byte {
	p := [][]byte{[]byte(supi), abba}
	l := []int{len(supi), len(abba)}
	return KDFBytes(kSeaf, 0x6D, p, l)
}

// DeriveKgnb derives Kgnb from Kamf (TS 33.501)
func DeriveKgnb(kAmf []byte, ulCount uint32, accessType byte) []byte {
	p := [][]byte{
		{byte(ulCount >> 24), byte(ulCount >> 16), byte(ulCount >> 8), byte(ulCount)},
		{accessType},
	}
	l := []int{4, 1}
	return KDFBytes(kAmf, 0x6E, p, l)
}

// DeriveKnas derives KnasInt or KnasEnc from Kamf (TS 33.501)
// algType: 1 for Encryption, 2 for Integrity
// algId: 1 for NEA1/NIA1, 2 for NEA2/NIA2, etc.
func DeriveKnas(kAmf []byte, algType byte, algId byte) []byte {
	p := [][]byte{{algType}, {algId}}
	l := []int{1, 1}
	k := KDFBytes(kAmf, 0x69, p, l)
	// Return only the last 16 bytes (128 bits)
	return k[16:]
}
