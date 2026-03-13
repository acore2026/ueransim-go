package utils

// BitBuffer provides bit-level access to a byte slice.
type BitBuffer struct {
	data  []byte
	index int // bit index
}

func NewBitBuffer(data []byte) *BitBuffer {
	return &BitBuffer{data: data, index: 0}
}

func (b *BitBuffer) Seek(index int) {
	b.index = index
}

func (b *BitBuffer) CurrentIndex() int {
	return b.index
}

func (b *BitBuffer) Peek() int {
	octetIndex := b.index / 8
	bitIndex := b.index % 8
	if octetIndex >= len(b.data) {
		return 0
	}
	return int((b.data[octetIndex] >> (7 - bitIndex)) & 0x01)
}

func (b *BitBuffer) Read() int {
	bit := b.Peek()
	b.index++
	return bit
}

func (b *BitBuffer) ReadBits(len int) int {
	i := 0
	for j := 0; j < len; j++ {
		i <<= 1
		i |= b.Read()
	}
	return i
}

func (b *BitBuffer) ReadUint64(len int) uint64 {
	var i uint64
	for j := 0; j < len; j++ {
		i <<= 1
		i |= uint64(b.Read())
	}
	return i
}

func (b *BitBuffer) Write(bit bool) {
	octetIndex := b.index / 8
	bitIndex := b.index % 8

	// Extend data if needed
	if octetIndex >= len(b.data) {
		needed := octetIndex - len(b.data) + 1
		b.data = append(b.data, make([]byte, needed)...)
	}

	if bit {
		b.data[octetIndex] |= (1 << (7 - bitIndex))
	} else {
		b.data[octetIndex] &= ^(1 << (7 - bitIndex))
	}
	b.index++
}

func (b *BitBuffer) WriteBits(value int, length int) {
	for i := 0; i < length; i++ {
		b.Write(((value >> (length - 1 - i)) & 0x01) == 1)
	}
}

func (b *BitBuffer) WriteUint64(value uint64, length int) {
	for i := 0; i < length; i++ {
		b.Write(((value >> (length - 1 - i)) & 0x01) == 1)
	}
}

func (b *BitBuffer) Data() []byte {
	return b.data
}

// BitString mirrors the BitString utility in C++.
type BitString struct {
	buf *BitBuffer
}

func NewBitString() *BitString {
	return &BitString{buf: NewBitBuffer(nil)}
}

func (s *BitString) Write(bit bool) {
	s.buf.Write(bit)
}

func (s *BitString) WriteBits(value int, length int) {
	s.buf.WriteBits(value, length)
}

func (s *BitString) BitLength() int {
	return s.buf.CurrentIndex()
}

func (s *BitString) OctetLength() int {
	return (s.BitLength() + 7) / 8
}

func (s *BitString) Data() []byte {
	return s.buf.Data()
}
