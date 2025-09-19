package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssembler(t *testing.T) {
	assemblyCode := `
main:
PUSH4 5
JUMP squared

squared:
DUP1
MULT
PRINT
`
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.asm")
	err := os.WriteFile(filePath, []byte(assemblyCode), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tokenizer, err := NewTokenizer(filePath)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := tokenizer.Tokenize()
	if err != nil {
		t.Fatal(err)
	}

	assembler := NewAssembler(tokens)
	chunk, err := assembler.Assemble()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(chunk)
	// c := NewChunk()
	// c.EmitByte(byte(OP_PUSH4))
	// c.EmitInt32(int32(2))
	// c.EmitByte(byte(OP_PUSH4))
	// c.EmitInt32(int32(3))
	// c.EmitByte(byte(OP_ADD))
	// c.EmitByte(byte(OP_PRINT))

	// assert.True(t, reflect.DeepEqual(chunk.Bytecode, c.Bytecode))

	// if !bytes.Equal(chunk.Bytecode, c.Bytecode) {
	// 	t.Errorf("Expected: %v\nGot:%v\n", chunk.Bytecode, c.Bytecode)
	// }

	vm := NewVM(chunk)
	err = vm.Run()
	assert.NoError(t, err)
}
