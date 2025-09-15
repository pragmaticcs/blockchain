package vm

import (
	"encoding/binary"
)

type ValueType byte

const (
	ValueTypeInt ValueType = iota
	ValueTypeBool
)

const (
	ValueIntSize  = 4
	ValueBoolSize = 1
)

type Value struct {
	Type ValueType
	Data []byte
}

type ValueInt struct {
	Type ValueType
	Int  int32
}

func NewValueInt(i int32) (Value, error) {
	data := [ValueIntSize]byte{}
	_, err := binary.Encode(data[:], binary.BigEndian, i)
	if err != nil {
		return Value{}, err
	}
	return Value{
		Type: ValueTypeInt,
		Data: data[:],
	}, nil
}

func NewValueBool(b bool) (Value, error) {
	data := [ValueBoolSize]byte{}
	_, err := binary.Encode(data[:], binary.BigEndian, b)
	if err != nil {
		return Value{}, err
	}
	return Value{
		Type: ValueTypeInt,
		Data: data[:],
	}, nil
}

func (v Value) VisitAsInt() (int, int32, error) {
	var value int32
	n, err := binary.Decode(v.Data, binary.BigEndian, &value)
	return n, value, err
}

func (v Value) VisitAsBool() (int, bool, error) {
	var value bool
	n, err := binary.Decode(v.Data, binary.BigEndian, &value)
	return n, value, err
}

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

func (c *Chunk) EmitValue(val Value) {
	c.Bytecode = append(c.Bytecode, byte(val.Type))
	c.Bytecode = append(c.Bytecode, val.Data...)
}
