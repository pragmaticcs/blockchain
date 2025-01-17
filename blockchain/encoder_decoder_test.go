package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGobEncoderDecoder(t *testing.T) {
	difficulty := 16
	bch, err := NewBlockChain(difficulty)
	if err != nil {
		t.Error(err)
	}

	coinbase, err := NewCoinbase(1000000, []byte("my public key"), []byte("my signature"))
	if err != nil {
		t.Error(err)
	}
	block := bch.AddBlock([]*Transaction{coinbase})

	encoderDecoder := GobEncoderDecoder{}
	encoding, err := encoderDecoder.EncodeBlock(block)
	if err != nil {
		t.Error(err)
	}

	decoding, err := encoderDecoder.DecodeBlock(encoding)
	if err != nil {
		t.Error(err)
	}
	assert.EqualValues(t, block, decoding)
}
