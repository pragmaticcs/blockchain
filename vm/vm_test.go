package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVM(t *testing.T) {
	c := NewChunk()
	c.EmitByte(byte(OP_PUSH4))
	c.EmitInt32(int32(2))
	c.EmitByte(byte(OP_PUSH4))
	c.EmitInt32(int32(3))
	c.EmitByte(byte(OP_ADD))
	c.EmitByte(byte(OP_PRINT))

	vm := NewVM(c)
	err := vm.Run()
	assert.NoError(t, err)
}
