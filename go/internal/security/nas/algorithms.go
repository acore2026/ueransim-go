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

	// IV: 
	// Count (32 bits)
	// Bearer (5 bits)
	// Direction (1 bit)
	// 0...0 (2 bits)
	// 0...0 (64 bits)
	// Note: 3GPP IV for CTR usually places these in the high bits.
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
	// Message prefix matching C++ bits::Ranged8
	// m.appendOctet4(count);
	// m.appendOctet(bits::Ranged8({{5, bearer}, {1, direction}, {2, 0}}));
	// m.appendOctet3(0);
	msg := make([]byte, 8+len(data))
	binary.BigEndian.PutUint32(msg[0:4], count)
	
	// bits::Ranged8 ensures the bits are shifted correctly.
	// {{5, bearer}, {1, direction}, {2, 0}}
	// bearer: bits 7-3
	// direction: bit 2
	// padding: bits 1-0
	msg[4] = (bearer << 3) | (direction << 2)
	
	copy(msg[8:], data)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mac, err := cmac.New(block)
	if err != nil {
		return nil, err
	}

	mac.Write(msg)
	fullMac := mac.Sum(nil)
	
	// Return the first 4 bytes as the MAC
	return fullMac[0:4], nil
}
