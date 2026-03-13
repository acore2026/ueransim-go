package utils

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// Buffer is a flexible byte slice with utility methods for 3GPP protocol encoding.
// It mirrors the functionality of OctetString in the C++ version.
type Buffer struct {
	data []byte
}

func NewBuffer(data []byte) *Buffer {
	return &Buffer{data: data}
}

func NewEmptyBuffer() *Buffer {
	return &Buffer{data: make([]byte, 0)}
}

func FromHex(h string) (*Buffer, error) {
	d, err := hex.DecodeString(h)
	if err != nil {
		return nil, err
	}
	return &Buffer{data: d}, nil
}

func (b *Buffer) Data() []byte {
	return b.data
}

func (b *Buffer) Len() int {
	return len(b.data)
}

func (b *Buffer) AppendByte(v byte) {
	b.data = append(b.data, v)
}

func (b *Buffer) AppendUint16(v uint16) {
	temp := make([]byte, 2)
	binary.BigEndian.PutUint16(temp, v)
	b.data = append(b.data, temp...)
}

func (b *Buffer) AppendUint24(v uint32) {
	b.data = append(b.data, byte(v>>16), byte(v>>8), byte(v))
}

func (b *Buffer) AppendUint32(v uint32) {
	temp := make([]byte, 4)
	binary.BigEndian.PutUint32(temp, v)
	b.data = append(b.data, temp...)
}

func (b *Buffer) AppendUint64(v uint64) {
	temp := make([]byte, 8)
	binary.BigEndian.PutUint64(temp, v)
	b.data = append(b.data, temp...)
}

func (b *Buffer) Append(other *Buffer) {
	if other != nil {
		b.data = append(b.data, other.data...)
	}
}

func (b *Buffer) AppendBytes(v []byte) {
	b.data = append(b.data, v...)
}

func (b *Buffer) GetByte(index int) byte {
	return b.data[index]
}

func (b *Buffer) GetUint16(index int) uint16 {
	return binary.BigEndian.Uint16(b.data[index : index+2])
}

func (b *Buffer) GetUint32(index int) uint32 {
	return binary.BigEndian.Uint32(b.data[index : index+4])
}

func (b *Buffer) Hex() string {
	return hex.EncodeToString(b.data)
}

func (b *Buffer) String() string {
	return b.Hex()
}

func (b *Buffer) Copy() *Buffer {
	d := make([]byte, len(b.data))
	copy(d, b.data)
	return &Buffer{data: d}
}

func (b *Buffer) SubCopy(index int, length int) (*Buffer, error) {
	if index < 0 || index+length > len(b.data) {
		return nil, fmt.Errorf("out of bounds")
	}
	d := make([]byte, length)
	copy(d, b.data[index:index+length])
	return &Buffer{data: d}, nil
}
