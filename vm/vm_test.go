package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVM(t *testing.T) {
	c := NewChunk()
	c.EmitByte(byte(OP_PUSH))
	val, _ := NewValueInt(1)
	c.EmitValue(val)
	c.EmitByte(byte(OP_PUSH))
	val, _ = NewValueInt(2)
	c.EmitValue(val)
	c.EmitByte(byte(OP_ADD))
	c.EmitByte(byte(OP_PRINT))

	vm := NewVM(c)
	err := vm.Run()
	assert.NoError(t, err)
}
