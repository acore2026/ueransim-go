package utils

import "fmt"

func EncodePlmn(mcc, mnc string) []byte {
	if len(mcc) != 3 {
		panic(fmt.Sprintf("invalid MCC length: %d", len(mcc)))
	}
	if len(mnc) != 2 && len(mnc) != 3 {
		panic(fmt.Sprintf("invalid MNC length: %d", len(mnc)))
	}

	res := make([]byte, 3)
	// Octet 1: MCC2 | MCC1
	res[0] = (mcc[1]-'0')<<4 | (mcc[0]-'0')
	
	// Octet 2: MNC3 | MCC3 (if 3-digit MNC) or 0xF | MCC3 (if 2-digit MNC)
	if len(mnc) == 3 {
		res[1] = (mnc[2]-'0')<<4 | (mcc[2]-'0')
	} else {
		res[1] = 0xF0 | (mcc[2]-'0')
	}
	
	// Octet 3: MNC2 | MNC1
	res[2] = (mnc[1]-'0')<<4 | (mnc[0]-'0')
	
	return res
}

func EncodeBcd(s string) []byte {
	res := make([]byte, (len(s)+1)/2)
	for i := 0; i < len(s); i++ {
		val := s[i] - '0'
		if i%2 == 0 {
			res[i/2] |= val
		} else {
			res[i/2] |= (val << 4)
		}
	}
	// If odd length, fill last nibble with 0xF
	if len(s)%2 != 0 {
		res[len(res)-1] |= 0xF0
	}
	return res
}
