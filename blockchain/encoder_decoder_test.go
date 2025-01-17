package blockchain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGobEncoderDecoder(t *testing.T) {
	difficulty := 16
	bch := NewBlockChain(difficulty)
	block := bch.AddBlock([]byte("data"))

	encoderDecoder := GobEncoderDecoder{}
	encoding, err := encoderDecoder.EncodeBlock(block)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Gob Encoding: ", encoding)

	decoding, err := encoderDecoder.DecodeBlock(encoding)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, block, decoding)
}
