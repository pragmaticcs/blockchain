package vm

import "encoding/binary"

type Chunk struct {
	Bytecode []byte
}

func NewChunk() *Chunk {
	return &Chunk{
		Bytecode: make([]byte, 0),
	}
}

func (c *Chunk) EmitByte(b byte) {
	c.Bytecode = append(c.Bytecode, b)
}

func (c *Chunk) EmitBytes(b1, b2 byte) {
	c.Bytecode = append(c.Bytecode, b1)
	c.Bytecode = append(c.Bytecode, b2)
}

func (c *Chunk) EmitInt32(val int32) {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, uint32(val))
	c.Bytecode = append(c.Bytecode, bytes...)
}

func (c *Chunk) EmitInt64(val int64) {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(val))
	c.Bytecode = append(c.Bytecode, bytes...)
}
