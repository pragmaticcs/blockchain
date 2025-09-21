package vm

import (
	"fmt"
	"strconv"
)

type OpCode byte

const (
	// Stack manipulation
	OP_PUSH1 OpCode = iota
	OP_PUSH4
	OP_PUSH8
	OP_POP
	OP_DUP1
	OP_DUP2

	// Arithmetic
	OP_ADD
	OP_SUB
	OP_MULT
	OP_DIV

	// Logic
	OP_EQ
	OP_GT
	OP_LT
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

var tokenPushLengths = map[TokenType]int{
	TOKEN_PUSH1: 1,
	TOKEN_PUSH4: 4,
	TOKEN_PUSH8: 8,
}

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
			}
		case TOKEN_PUSH1, TOKEN_PUSH4, TOKEN_PUSH8:
			if a.peek(0).Type != TOKEN_NUMBER {
				return fmt.Errorf("[%d] expected number after PUSH<N>", token.Line)
			}
			a.consume()
			locCounter += 1 + tokenPushLengths[token.Type]
		case TOKEN_JUMP, TOKEN_JUMPI:
			if a.peek(0).Type != TOKEN_IDENTIFIER {
				return fmt.Errorf("[%d] expected label after %s", token.Line, token.Lexeme)
			}
			a.consume()
			locCounter += 1 + 4 + 1
		case TOKEN_NUMBER:
			return fmt.Errorf("[%d] unexpected number", token.Line)
		case TOKEN_ADD, TOKEN_DIV, TOKEN_MULT, TOKEN_SUB, TOKEN_POP, TOKEN_PRINT, TOKEN_DUP1:
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

func (a *Assembler) SecondPass() (*Chunk, error) {
	a.currToken = 0 // reset from the first pass
Loop:
	for {
		token := a.consume()
		switch token.Type {
		case TOKEN_IDENTIFIER:
			if a.peek(0).Type == TOKEN_COLON {
				a.consume() // consume colon
			} else {
				return nil, fmt.Errorf("[%d] unexpected identifier '%s'", token.Line, token.Lexeme)
			}
		case TOKEN_COLON:
			return nil, fmt.Errorf("[%d] unexpected `:`", token.Line)
		case TOKEN_PUSH4:
			a.chunk.EmitByte(byte(OP_PUSH4))
			numToken := a.consume() // don't need to check if it is a number
			// cause we already checked it in the first pass
			val, _ := strconv.Atoi(numToken.Lexeme)
			a.chunk.EmitInt32(int32(val))
		case TOKEN_PUSH1:
			numToken := a.consume()
			val, _ := strconv.Atoi(numToken.Lexeme)
			a.chunk.EmitByte(byte(val))
		case TOKEN_JUMP, TOKEN_JUMPI:
			labelToken := a.consume()
			address, ok := a.symbolicTable[labelToken.Lexeme]
			if !ok {
				return nil, fmt.Errorf("[%d] undefined label '%s'", labelToken.Line, labelToken.Lexeme)
			}
			a.chunk.EmitByte(byte(OP_PUSH4))
			a.chunk.EmitInt32(int32(address))

			if token.Type == TOKEN_JUMP {
				a.chunk.EmitByte(byte(OP_JUMP))
			} else {
				a.chunk.EmitByte(byte(OP_JUMPI))
			}
		case TOKEN_POP:
			a.chunk.EmitByte(byte(OP_POP))
		case TOKEN_ADD:
			a.chunk.EmitByte(byte(OP_ADD))
		case TOKEN_SUB:
			a.chunk.EmitByte(byte(OP_SUB))
		case TOKEN_MULT:
			a.chunk.EmitByte(byte(OP_MULT))
		case TOKEN_DIV:
			a.chunk.EmitByte(byte(OP_DIV))
		case TOKEN_DUP1:
			a.chunk.EmitByte(byte(OP_DUP1))
		case TOKEN_PRINT:
			a.chunk.EmitByte(byte(OP_PRINT))
		case TOKEN_EOF:
			break Loop
		}
	}
	a.chunk.EmitByte(byte(OP_END))
	return a.chunk, nil
}
