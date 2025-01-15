package blockchain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockChain(t *testing.T) {
	difficulty := 16
	bch := NewBlockChain(difficulty)
	for i := 1; i < 10; i++ {
		block := bch.AddBlock([]byte(fmt.Sprintf("Block [%d]", i)))
		pow := NewPOW(block, uint(difficulty))
		assert.Equal(t, pow.Validate(), true)
		t.Logf("Mined and Validated block %d. Nonce = %d\n", i, block.Nonce)
	}
}
