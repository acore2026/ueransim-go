package rls

import (
	"encoding/binary"
	"fmt"

	"github.com/acore2026/ueransim-go/internal/utils"
)

type MessageType uint8

const (
	RESERVED             MessageType = 0
	HEARTBEAT            MessageType = 4
	HEARTBEAT_ACK        MessageType = 5
	PDU_TRANSMISSION     MessageType = 6
	PDU_TRANSMISSION_ACK MessageType = 7
)

type PduType uint8

const (
	PDU_TYPE_RESERVED PduType = 0
	PDU_TYPE_RRC      PduType = 1
	PduTypeData       PduType = 2
)

type RlsMessage struct {
	MsgType MessageType
	Sti     uint64
	
	// Heartbeat
	SimPos [3]int32
	
	// Heartbeat Ack
	Dbm int32
	
	// Pdu Transmission
	PduType   PduType
	PduId     uint32
	Payload   uint32
	Pdu       []byte
	
	// Pdu Transmission Ack
	PduIds []uint32
}

func BuildSimpleRrc(nasPdu []byte) []byte {
	b := utils.NewEmptyBuffer()
	b.AppendByte(0x01) // Type: Container
	b.AppendUint32(uint32(len(nasPdu)))
	b.AppendBytes(nasPdu)
	return b.Data()
}

func (m *RlsMessage) Encode() ([]byte, error) {
	b := utils.NewEmptyBuffer()
	
	b.AppendByte(0x03) // Old RLS compatibility
	
	// Major, Minor, Patch (Hardcoded for now, should match src/utils/constants.hpp)
	b.AppendByte(3) // Major
	b.AppendByte(2) // Minor
	b.AppendByte(7) // Patch
	
	b.AppendByte(byte(m.MsgType))
	b.AppendUint64(m.Sti)
	
	switch m.MsgType {
	case HEARTBEAT:
		b.AppendUint32(uint32(m.SimPos[0]))
		b.AppendUint32(uint32(m.SimPos[1]))
		b.AppendUint32(uint32(m.SimPos[2]))
	case HEARTBEAT_ACK:
		b.AppendUint32(uint32(m.Dbm))
	case PDU_TRANSMISSION:
		b.AppendByte(byte(m.PduType))
		b.AppendUint32(m.PduId)
		b.AppendUint32(m.Payload)
		b.AppendUint32(uint32(len(m.Pdu)))
		b.AppendBytes(m.Pdu)
	case PDU_TRANSMISSION_ACK:
		b.AppendUint32(uint32(len(m.PduIds)))
		for _, id := range m.PduIds {
			b.AppendUint32(id)
		}
	}
	
	return b.Data(), nil
}

func Decode(data []byte) (*RlsMessage, error) {
	if len(data) < 13 {
		return nil, fmt.Errorf("data too short")
	}
	
	if data[0] != 0x03 {
		return nil, fmt.Errorf("invalid RLS compatibility byte")
	}
	
	// Skip Major, Minor, Patch for now
	msgType := MessageType(data[4])
	sti := binary.BigEndian.Uint64(data[5:13])
	
	m := &RlsMessage{
		MsgType: msgType,
		Sti:     sti,
	}
	
	offset := 13
	switch msgType {
	case HEARTBEAT:
		m.SimPos[0] = int32(binary.BigEndian.Uint32(data[offset : offset+4]))
		m.SimPos[1] = int32(binary.BigEndian.Uint32(data[offset+4 : offset+8]))
		m.SimPos[2] = int32(binary.BigEndian.Uint32(data[offset+8 : offset+12]))
	case HEARTBEAT_ACK:
		m.Dbm = int32(binary.BigEndian.Uint32(data[offset : offset+4]))
	case PDU_TRANSMISSION:
		m.PduType = PduType(data[offset])
		m.PduId = binary.BigEndian.Uint32(data[offset+1 : offset+5])
		m.Payload = binary.BigEndian.Uint32(data[offset+5 : offset+9])
		pduLen := binary.BigEndian.Uint32(data[offset+9 : offset+13])
		if offset+13+int(pduLen) > len(data) {
			return nil, fmt.Errorf("PDU length mismatch")
		}
		m.Pdu = data[offset+13 : offset+13+int(pduLen)]
	case PDU_TRANSMISSION_ACK:
		count := binary.BigEndian.Uint32(data[offset : offset+4])
		m.PduIds = make([]uint32, count)
		for i := 0; i < int(count); i++ {
			m.PduIds[i] = binary.BigEndian.Uint32(data[offset+4+i*4 : offset+8+i*4])
		}
	}
	
	return m, nil
}
