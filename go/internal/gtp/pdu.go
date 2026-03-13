package gtp

import (
	"encoding/binary"
	"fmt"

	"github.com/acore2026/ueransim-go/internal/utils"
)

// GTP Message Types
const (
	MT_ECHO_REQUEST      uint8 = 1
	MT_ECHO_RESPONSE     uint8 = 2
	MT_ERROR_INDICATION  uint8 = 26
	MT_END_MARKER        uint8 = 254
	MT_G_PDU             uint8 = 255
)

// GtpMessage represents a GTP-U PDU.
type GtpMessage struct {
	IFlag   bool
	SFlag   bool
	PNFlag  bool
	MsgType uint8
	Teid    uint32
	Seq     uint16
	NPduNum uint8
	Payload []byte
}

func (m *GtpMessage) Encode() ([]byte, error) {
	b := utils.NewEmptyBuffer()
	
	// Flags: Version(3 bits)=1, PT(1 bit)=1, Spare(1 bit)=0, E(1 bit), S(1 bit), PN(1 bit)
	var flags byte = 0x30 // Version 1, PT 1
	if m.SFlag { flags |= 0x02 }
	if m.PNFlag { flags |= 0x01 }
	// Extension headers not implemented yet, so E flag is 0
	
	b.AppendByte(flags)
	b.AppendByte(m.MsgType)
	
	// Length: will be updated later
	b.AppendUint16(0)
	b.AppendUint32(m.Teid)
	
	if m.SFlag || m.PNFlag {
		b.AppendUint16(m.Seq)
		b.AppendByte(m.NPduNum)
		b.AppendByte(0) // Next Extension Header Type
	}
	
	b.AppendBytes(m.Payload)
	
	data := b.Data()
	// Update Length (bytes 2-3)
	// Length is total length of payload + mandatory fields after TEID
	var gtpLen uint16
	if m.SFlag || m.PNFlag {
		gtpLen = uint16(len(m.Payload) + 4)
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
		IFlag:   (flags & 0x04) != 0,
		SFlag:   (flags & 0x02) != 0,
		PNFlag:  (flags & 0x01) != 0,
		MsgType: data[1],
		Teid:    binary.BigEndian.Uint32(data[4:8]),
	}
	
	payloadOffset := 8
	if m.IFlag || m.SFlag || m.PNFlag {
		if len(data) < 12 {
			return nil, fmt.Errorf("GTP data too short for flags")
		}
		m.Seq = binary.BigEndian.Uint16(data[8:10])
		m.NPduNum = data[10]
		// Byte 11 is Next Extension Header Type
		payloadOffset = 12
	}
	
	m.Payload = data[payloadOffset:]
	return m, nil
}
