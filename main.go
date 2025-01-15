package main

import (
	"fmt"

	"github.com/pragmaticcs/blockchain/blockchain"
)

func main() {
	difficulty := 12
	bch := blockchain.NewBlockChain(difficulty)
	for i := 1; i <= 5; i++ {
		block := bch.AddBlock([]byte(fmt.Sprintf("Block [%d]", i)))
		pow := blockchain.NewPOW(block, uint(difficulty))
		if !pow.Validate() {
			fmt.Printf("Couldn't validate block %d\n", i)
		}
		fmt.Printf("Mined block %d. Nonce = %d\n", i, block.Nonce)
	}

}
