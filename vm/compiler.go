package vm

type OpCode byte

const (
	// Stack manipulation
	OP_PUSH OpCode = iota
	OP_POP
	OP_DUP

	// Arithmetic
	OP_ADD
	OP_SUB
	OP_MULT
	OP_DIV

	// Logic
	OP_AND
	OP_OR

	// Control Flow
	OP_JUMP
	OP_JUMPI

	// Memory
	MSTORE
	MLOAD
	SSTORE
	SLOAD

	// Helper
	OP_PRINT
	OP_END
)

type Compiler struct {
	Chunk *Chunk
}
