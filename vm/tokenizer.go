package vm

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

type TokenType byte

const (
	TOKEN_COLON TokenType = iota

	TOKEN_IDENTIFIER
	TOKEN_NUMBER

	TOKEN_PUSH1
	TOKEN_PUSH4
	TOKEN_PUSH8
	TOKEN_POP
	TOKEN_ADD
	TOKEN_SUB
	TOKEN_DIV
	TOKEN_MULT
	TOKEN_PRINT

	TOKEN_ERROR
	TOKEN_EOF
)

var keywords = map[string]TokenType{
	"PUSH1": TOKEN_PUSH1,
	"PUSH4": TOKEN_PUSH4,
	"PUSH8": TOKEN_PUSH8,
	"POP":   TOKEN_POP,
	"ADD":   TOKEN_ADD,
	"SUB":   TOKEN_SUB,
	"DIV":   TOKEN_DIV,
	"MULT":  TOKEN_MULT,
	"PRINT": TOKEN_PRINT,
}

type Token struct {
	Type   TokenType
	Lexeme string
	Line   int
}

func NewToken(t TokenType, lexeme string, line int) Token {
	return Token{
		Type:   t,
		Lexeme: lexeme,
		Line:   line,
	}
}

type Tokenizer struct {
	fileData []byte
	start    int
	current  int
	line     int
}

func NewTokenizer(path string) (*Tokenizer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &Tokenizer{
		fileData: data,
		start:    0,
		current:  0,
		line:     1,
	}, nil
}

func (t *Tokenizer) Tokenize() ([]Token, error) {
	var tokens []Token
	for !t.isAtEnd() {
		t.start = t.current
		token := t.scanToken()

		if token.Type == TOKEN_ERROR {
			return nil, fmt.Errorf("[line %d] error: %s", token.Line, token.Lexeme)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (t *Tokenizer) scanToken() Token {
	t.skipWhitespaceAndComments()

	if t.isAtEnd() {
		return NewToken(TOKEN_EOF, "", t.line)
	}

	t.start = t.current
	c := t.consume()

	if unicode.IsLetter(c) {
		return t.tokenizeIdentifier()
	}
	if unicode.IsDigit(c) {
		return t.tokenizeNumber()
	}

	if c == ':' {
		return t.makeToken(TOKEN_COLON)
	}

	return t.errorToken(fmt.Sprintf("Unexpected character: %c", c))
}

func (t *Tokenizer) tokenizeIdentifier() Token {
	for !t.isAtEnd() && (isAlphaNumeric(t.peek()) || t.peek() == '_') {
		t.consume()
	}

	text := string(t.fileData[t.start:t.current])

	tokenType, ok := keywords[strings.ToUpper(text)]
	if ok {
		return t.makeToken(tokenType)
	}

	return t.makeToken(TOKEN_IDENTIFIER)
}

func (t *Tokenizer) tokenizeNumber() Token {
	for !t.isAtEnd() && unicode.IsDigit(t.peek()) {
		t.consume()
	}
	return t.makeToken(TOKEN_NUMBER)
}

func (t *Tokenizer) skipWhitespaceAndComments() {
	for !t.isAtEnd() {
		c := t.peek()
		switch c {
		case '\n':
			t.line++
			t.consume()
		case ' ', '\r', '\t':
			t.consume()
		case '#':
			for !t.isAtEnd() && t.peek() != '\n' {
				t.consume()
			}
		default:
			return
		}
	}
}

func (t *Tokenizer) isAtEnd() bool {
	return t.current >= len(t.fileData)
}

func (t *Tokenizer) consume() rune {
	defer func() { t.current++ }()
	return t.peek()
}

func (t *Tokenizer) peek() rune {
	if t.isAtEnd() {
		return '\x00' // For EOF
	}
	return rune(t.fileData[t.current])
}

func (t *Tokenizer) makeToken(tokenType TokenType) Token {
	lexeme := string(t.fileData[t.start:t.current])
	return NewToken(tokenType, lexeme, t.line)
}

func (t *Tokenizer) errorToken(message string) Token {
	return NewToken(TOKEN_ERROR, message, t.line)
}

func isAlphaNumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
