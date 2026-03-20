package gtp

import (
	"encoding/binary"
	"fmt"

	"github.com/acore2026/ueransim-go/internal/utils"
)

// GTP Message Types
const (
	MT_ECHO_REQUEST     uint8 = 1
	MT_ECHO_RESPONSE    uint8 = 2
	MT_ERROR_INDICATION uint8 = 26
	MT_END_MARKER       uint8 = 254
	MT_G_PDU            uint8 = 255
)

// GtpMessage represents a GTP-U PDU.
type GtpMessage struct {
	EFlag   bool
	SFlag   bool
	PNFlag  bool
	MsgType uint8
	Teid    uint32
	Seq     uint16
	NPduNum uint8
	QFI     *uint8
	Payload []byte
}

func (m *GtpMessage) Encode() ([]byte, error) {
	b := utils.NewEmptyBuffer()

	// Flags: Version(3 bits)=1, PT(1 bit)=1, Spare(1 bit)=0, E(1 bit), S(1 bit), PN(1 bit)
	var flags byte = 0x30 // Version 1, PT 1
	if m.EFlag {
		flags |= 0x04
	}
	if m.SFlag {
		flags |= 0x02
	}
	if m.PNFlag {
		flags |= 0x01
	}

	b.AppendByte(flags)
	b.AppendByte(m.MsgType)

	// Length: will be updated later
	b.AppendUint16(0)
	b.AppendUint32(m.Teid)

	if m.EFlag || m.SFlag || m.PNFlag {
		b.AppendUint16(m.Seq)
		b.AppendByte(m.NPduNum)
		if m.QFI != nil {
			// PDU Session Container extension header for N3 uplink.
			b.AppendByte(0x85)
			b.AppendByte(1)
			b.AppendByte(0x10) // UL PDU Session Information, all optional fields absent.
			b.AppendByte(*m.QFI & 0x3f)
		}
		b.AppendByte(0) // No more extension headers.
	}

	b.AppendBytes(m.Payload)

	data := b.Data()
	// Update Length (bytes 2-3)
	// Length is total length of payload + mandatory fields after TEID
	var gtpLen uint16
	if m.EFlag || m.SFlag || m.PNFlag {
		gtpLen = uint16(len(m.Payload) + 4)
		if m.QFI != nil {
			gtpLen += 3 // type + len + 2-byte UL PDU Session Information payload + trailing next-header already counted below?
			gtpLen++    // final next-extension-header octet
		} else {
			gtpLen++ // final next-extension-header octet
		}
	} else {
		gtpLen = uint16(len(m.Payload))
	}
	binary.BigEndian.PutUint16(data[2:4], gtpLen)

	return data, nil
}

func Decode(data []byte) (*GtpMessage, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("GTP data too short")
	}

	flags := data[0]
	version := (flags >> 5) & 0x07
	if version != 1 {
		return nil, fmt.Errorf("unsupported GTP version: %d", version)
	}

	m := &GtpMessage{
		EFlag:   (flags & 0x04) != 0,
		SFlag:   (flags & 0x02) != 0,
		PNFlag:  (flags & 0x01) != 0,
		MsgType: data[1],
		Teid:    binary.BigEndian.Uint32(data[4:8]),
	}

	payloadOffset := 8
	if m.EFlag || m.SFlag || m.PNFlag {
		if len(data) < 12 {
			return nil, fmt.Errorf("GTP data too short for flags")
		}
		m.Seq = binary.BigEndian.Uint16(data[8:10])
		m.NPduNum = data[10]
		nextExtHeaderType := data[11]
		payloadOffset = 12
		if m.EFlag {
			for nextExtHeaderType != 0 {
				if len(data) < payloadOffset+2 {
					return nil, fmt.Errorf("GTP extension header too short")
				}
				extLenUnits := int(data[payloadOffset])
				extStart := payloadOffset + 1
				extBytes := 4*extLenUnits - 2
				extEnd := extStart + extBytes
				if extLenUnits <= 0 || len(data) < extEnd+1 {
					return nil, fmt.Errorf("invalid GTP extension header length")
				}
				if nextExtHeaderType == 0x85 && extBytes >= 2 {
					qfi := data[extStart+1] & 0x3f
					m.QFI = &qfi
				}
				nextExtHeaderType = data[extEnd]
				payloadOffset = extEnd + 1
			}
		}
	}

	m.Payload = data[payloadOffset:]
	return m, nil
}
