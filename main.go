package main

import (
	"fmt"
	"log"

	"github.com/pragmaticcs/blockchain/blockchain"
)

func main() {
	difficulty := 12
	bch, err := blockchain.NewBlockChain(difficulty)
	if err != nil {
		log.Fatal(err)
	}
	for i := 1; i <= 5; i++ {
		block := bch.AddBlock([]*blockchain.Transaction{})
		pow := blockchain.NewPOW(block, uint(difficulty))
		if !pow.Validate() {
			fmt.Printf("Couldn't validate block %d\n", i)
		}
		fmt.Printf("Mined block %d. Nonce = %d\n", i, block.Nonce)
	}

}
