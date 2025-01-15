package blockchain

import (
	"log"
	"testing"
)

func TestBlockChain(t *testing.T) {
	difficulty := 16
	bch := NewBlockChain(difficulty)
	block := bch.AddBlock([]byte("data"))
	log.Println(block.Nonce)
}
