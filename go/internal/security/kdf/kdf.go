package kdf

import (
	"crypto/hmac"
	"crypto/sha256"
)

// KDF implements the 3GPP Key Derivation Function (TS 33.220)
// S = FC || P0 || L0 || P1 || L1 || ... || Pn || Ln
func KDF(key []byte, fc byte, p [][]byte, l []int) []byte {
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

	k := KDF(key, 0x6B, p, l)
	return k[16:]
}

// DeriveKseaf derives Kseaf from CK, IK (TS 33.501, Annex A.6)
func DeriveKausf(ck, ik []byte, snName string, sqnXorAk []byte) []byte {
	key := append(ck, ik...)
	p := [][]byte{[]byte(snName), sqnXorAk}
	l := []int{len(snName), len(sqnXorAk)}
	return KDF(key, 0x6A, p, l)
}

// DeriveKseaf derives Kseaf from Kausf (TS 33.501, Annex A.6)
func DeriveKseaf(kAusf []byte, snName string) []byte {
	p := [][]byte{[]byte(snName)}
	l := []int{len(snName)}
	return KDF(kAusf, 0x6C, p, l)
}

// DeriveKamf derives Kamf from Kseaf (TS 33.501, Annex A.7)
func DeriveKamf(kSeaf []byte, supi string, abba []byte) []byte {
	p := [][]byte{[]byte(supi), abba}
	l := []int{len(supi), len(abba)}
	return KDF(kSeaf, 0x6D, p, l)
}

// DeriveKgnb derives Kgnb from Kamf (TS 33.501)
func DeriveKgnb(kAmf []byte, ulCount uint32, accessType byte) []byte {
	p := [][]byte{
		{byte(ulCount >> 24), byte(ulCount >> 16), byte(ulCount >> 8), byte(ulCount)},
		{accessType},
	}
	l := []int{4, 1}
	return KDF(kAmf, 0x6E, p, l)
}

// DeriveKnas derives KnasInt or KnasEnc from Kamf (TS 33.501)
func DeriveKnas(kAmf []byte, algType byte, algId byte) []byte {
	p := [][]byte{{algType}, {algId}}
	l := []int{1, 1}
	k := KDF(kAmf, 0x69, p, l)
	// TS 33.501 Annex A.8: LSB 128 bits
	return k[16:]
}
