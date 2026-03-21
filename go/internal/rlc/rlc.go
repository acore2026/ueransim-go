package rlc

type Mode int

const (
	ModeTM Mode = iota
	ModeUM
	ModeAM
)

const (
	SIComplete uint8 = 0 // 00: Complete SDU
	SIFirst    uint8 = 1 // 01: First segment
	SILast     uint8 = 2 // 10: Last segment
	SIMiddle   uint8 = 3 // 11: Middle segment
)

// Simplified RLC-UM Header (1 byte)
// [SI (2 bits)][SN (6 bits)]
type UMHeader struct {
	SI uint8
	SN uint8
}

func (h *UMHeader) Encode() byte {
	return (h.SI << 6) | (h.SN & 0x3F)
}

func DecodeUMHeader(b byte) UMHeader {
	return UMHeader{
		SI: b >> 6,
		SN: b & 0x3F,
	}
}
