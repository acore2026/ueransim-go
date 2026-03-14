package milenage

import (
	"crypto/aes"
)

// Milenage implements 3GPP TS 35.206 Milenage algorithm.

type Milenage struct {
	k   []byte
	opc []byte
}

func NewMilenage(k, opc []byte) *Milenage {
	return &Milenage{
		k:   k,
		opc: opc,
	}
}

func aes128EncryptBlock(key, in []byte) []byte {
	block, _ := aes.NewCipher(key)
	out := make([]byte, 16)
	block.Encrypt(out, in)
	return out
}

func (m *Milenage) F1(rand, sqn, amf []byte) (macA, macS []byte) {
	/* tmp1 = TEMP = E_K(RAND XOR OP_C) */
	tmp1 := make([]byte, 16)
	for i := 0; i < 16; i++ {
		tmp1[i] = rand[i] ^ m.opc[i]
	}
	temp := aes128EncryptBlock(m.k, tmp1)

	/* in1 = SQN || AMF || SQN || AMF */
	in1 := make([]byte, 16)
	copy(in1[0:6], sqn)
	copy(in1[6:8], amf)
	copy(in1[8:14], sqn)
	copy(in1[14:16], amf)

	/* rotate (in1 XOR OP_C) by r1 (= 8 bytes) */
	tmp3 := make([]byte, 16)
	for i := 0; i < 16; i++ {
		tmp3[(i+8)%16] = in1[i] ^ m.opc[i]
	}
	/* XOR with TEMP */
	for i := 0; i < 16; i++ {
		tmp3[i] ^= temp[i]
	}

	/* out = E_K(tmp3) XOR OP_c */
	out := aes128EncryptBlock(m.k, tmp3)
	for i := 0; i < 16; i++ {
		out[i] ^= m.opc[i]
	}

	macA = make([]byte, 8)
	copy(macA, out[0:8])
	macS = make([]byte, 8)
	copy(macS, out[8:16])
	return
}

func (m *Milenage) F2345(rand []byte) (res, ck, ik, ak, akStar []byte) {
	/* tmp1 = TEMP = E_K(RAND XOR OP_C) */
	tmp1 := make([]byte, 16)
	for i := 0; i < 16; i++ {
		tmp1[i] = rand[i] ^ m.opc[i]
	}
	temp := aes128EncryptBlock(m.k, tmp1)

	// f2 and f5
	/* rotate (temp XOR OP_C) by r2 (= 0) */
	tmp2 := make([]byte, 16)
	for i := 0; i < 16; i++ {
		tmp2[i] = temp[i] ^ m.opc[i]
	}
	tmp2[15] ^= 1 // XOR c2
	out25 := aes128EncryptBlock(m.k, tmp2)
	for i := 0; i < 16; i++ {
		out25[i] ^= m.opc[i]
	}
	// f2 is bytes 8..15, f5 is bytes 0..5
	res = make([]byte, 8)
	copy(res, out25[8:16])
	ak = make([]byte, 6)
	copy(ak, out25[0:6])

	// f3 (ck)
	/* rotate (temp XOR OP_C) by r3 (= 4 bytes) */
	tmp3 := make([]byte, 16)
	for i := 0; i < 16; i++ {
		tmp3[(i+12)%16] = temp[i] ^ m.opc[i]
	}
	tmp3[15] ^= 2 // XOR c3
	out3 := aes128EncryptBlock(m.k, tmp3)
	ck = make([]byte, 16)
	for i := 0; i < 16; i++ {
		ck[i] = out3[i] ^ m.opc[i]
	}

	// f4 (ik)
	/* rotate (temp XOR OP_C) by r4 (= 8 bytes) */
	tmp4 := make([]byte, 16)
	for i := 0; i < 16; i++ {
		tmp4[(i+8)%16] = temp[i] ^ m.opc[i]
	}
	tmp4[15] ^= 4 // XOR c4
	out4 := aes128EncryptBlock(m.k, tmp4)
	ik = make([]byte, 16)
	for i := 0; i < 16; i++ {
		ik[i] = out4[i] ^ m.opc[i]
	}

	// f5* (akStar)
	/* rotate (temp XOR OP_C) by r5 (= 12 bytes) */
	tmp5 := make([]byte, 16)
	for i := 0; i < 16; i++ {
		tmp5[(i+4)%16] = temp[i] ^ m.opc[i]
	}
	tmp5[15] ^= 8 // XOR c5
	out5star := aes128EncryptBlock(m.k, tmp5)
	akStar = make([]byte, 6)
	for i := 0; i < 6; i++ {
		akStar[i] = out5star[i] ^ m.opc[i]
	}

	return
}

func GenerateOpC(k, op []byte) []byte {
	opc := aes128EncryptBlock(k, op)
	for i := 0; i < 16; i++ {
		opc[i] ^= op[i]
	}
	return opc
}
