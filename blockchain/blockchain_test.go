package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockChain(t *testing.T) {
	difficulty := 16
	bch, err := NewBlockChain(difficulty)
	if err != nil {
		t.Error(err)
	}
	for i := 1; i < 10; i++ {
		block := bch.AddBlock([]*Transaction{})
		pow := NewPOW(block, uint(difficulty))
		assert.Equal(t, pow.Validate(), true)
		t.Logf("Mined and Validated block %d. Nonce = %d\n", i, block.Nonce)
	}
}
