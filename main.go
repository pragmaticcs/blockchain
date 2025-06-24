package main

import (
	"fmt"
	"log"

	"github.com/pragmaticcs/blockchain/blockchain"
	"github.com/pragmaticcs/blockchain/wallet"
)

func main() {
	wallets := wallet.Wallets{
		Wallets: map[wallet.Address]*wallet.Wallet{},
	}
	bob, err := wallets.AddWallet()
	if err != nil {
		log.Fatal(err)
	}
	alice, err := wallets.AddWallet()
	if err != nil {
		log.Fatal(err)
	}

	wallet.SaveWalletsToFile(&wallets)

	bobsWallet, err := wallets.GetWallet(bob)
	if err != nil {
		log.Fatal(err)
	}
	// alicesWallet, _ := wallets.GetWallet(alice)
	genesisConf := blockchain.GenesisConfig{Amount: 50, To: bobsWallet.PublicKey, Sig: nil}

	difficulty := 12
	bch, err := blockchain.NewBlockChain(difficulty, genesisConf)

	if err != nil {
		log.Fatal(err)
	}

	firstTx, err := bch.NewTransaction(bob, alice, 10)
	if err != nil {
		log.Fatal(err)
	}
	block := bch.AddBlock([]*blockchain.Transaction{firstTx})
	pow := blockchain.NewPOW(block, uint(difficulty))
	if !pow.Validate() {
		log.Fatal("Couldn't validate block")
	}
	fmt.Printf("Mined block. Nonce = %d\n", block.Nonce)

	// for i := 1; i <= 5; i++ {
	// 	block := bch.AddBlock([]*blockchain.Transaction{})
	// 	pow := blockchain.NewPOW(block, uint(difficulty))
	// 	if !pow.Validate() {
	// 		fmt.Printf("Couldn't validate block %d\n", i)
	// 	}
	// 	fmt.Printf("Mined block %d. Nonce = %d\n", i, block.Nonce)
	// }
}
