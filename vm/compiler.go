package vm

type OpCode byte

const (
	OP_PUSH OpCode = iota
	// Pushes new value on top of the stack
	OP_POP
	// Pops the value of the stack

	OP_ADD
	OP_SUB
	OP_MULT
	OP_DIV

	OP_PRINT
	OP_END
)

type Compiler struct {
	Chunk *Chunk
}
