package vm

import "fmt"

type OpCode byte

const (
	// Stack manipulation
	OP_PUSH OpCode = iota
	OP_POP
	OP_DUP1
	OP_DUP2

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

type Assembler struct {
	symbolicTable map[string]int
	currToken     int
	tokens        []Token
	chunk         *Chunk
}

func NewAssembler(tokens []Token) *Assembler {
	return &Assembler{
		symbolicTable: make(map[string]int),
		currToken:     0,
		tokens:        tokens,
		chunk:         NewChunk(),
	}
}

func (a *Assembler) Assemble() (*Chunk, error) {
	err := a.FirstPass()
	if err != nil {
		return nil, err
	}
	return a.SecondPass()
}

func (a *Assembler) FirstPass() error {
	locCounter := 0

	for a.peek(0).Type != TOKEN_EOF {
		token := a.consume()
		switch token.Type {
		case TOKEN_IDENTIFIER:
			if a.peek(0).Type == TOKEN_COLON {
				a.consume() // consume colon
				a.symbolicTable[token.Lexeme] = locCounter

			} else {
				return fmt.Errorf("[%d] expected `:` after label", token.Line)
			}
		case TOKEN_PUSH:
			if a.peek(0).Type != TOKEN_NUMBER {
				return fmt.Errorf("[%d] expected number after PUSH", token.Line)
			}
			a.consume()
			locCounter += 5
		case TOKEN_ADD, TOKEN_DIV, TOKEN_MULT, TOKEN_SUB, TOKEN_POP, TOKEN_PRINT:
			locCounter += 1
		case TOKEN_EOF:
			return nil
		}
	}
	return nil
}

func (a *Assembler) consume() Token {
	defer func() { a.currToken++ }()
	return a.peek(0)
}

func (a *Assembler) peek(offset int) Token {
	if a.currToken+offset >= len(a.tokens) {
		return NewToken(TOKEN_EOF, "", -1)
	}
	return a.tokens[a.currToken+offset]
}

func (A *Assembler) SecondPass() (*Chunk, error) {
	return nil, nil
}
