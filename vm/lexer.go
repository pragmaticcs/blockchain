package vm

type TokenType byte

const (
	TOKEN_PUSH TokenType = iota
	TOKEN_POP
)

type Lexer struct {
}

func NewLexer() *Lexer {
	return &Lexer{}
}
