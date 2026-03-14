package nas

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"

	"github.com/aead/cmac"
)

// NEA2 (AES-CTR)
func NEA2(key []byte, count uint32, bearer byte, direction byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, 16)
	binary.BigEndian.PutUint32(iv[0:4], count)
	iv[4] = (bearer << 3) | (direction << 2)

	stream := cipher.NewCTR(block, iv)
	out := make([]byte, len(data))
	stream.XORKeyStream(out, data)
	return out, nil
}

// NIA2 (AES-CMAC)
func NIA2(key []byte, count uint32, bearer byte, direction byte, data []byte) ([]byte, error) {
	// TS 33.501 / TS 33.401 NIA2 (AES-CMAC)
	// Input: COUNT, BEARER, DIRECTION, MESSAGE
	// M = COUNT || BEARER || DIRECTION || 0...0 || MESSAGE
	
	m := make([]byte, 8+len(data))
	binary.BigEndian.PutUint32(m[0:4], count)
	m[4] = (bearer << 3) | (direction << 2)
	// m[5], m[6], m[7] are zero
	copy(m[8:], data)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mac, err := cmac.New(block)
	if err != nil {
		return nil, err
	}

	mac.Write(m)
	fullMac := mac.Sum(nil)
	
	return fullMac[0:4], nil
}
