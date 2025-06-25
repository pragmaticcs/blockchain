package main

import (
	"fmt"
	"log"

	"github.com/pragmaticcs/blockchain/blockchain"
	"github.com/pragmaticcs/blockchain/wallet"
)

func handle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	wallets := &wallet.Wallets{
		Wallets:      map[wallet.Address]*wallet.Wallet{},
		SaveFilePath: "wallets.json",
	}
	bob, err := wallets.AddWallet()
	handle(err)
	alice, err := wallets.AddWallet()
	handle(err)

	err = wallets.SaveWalletsToFile()
	handle(err)

	bobsWallet, err := wallets.GetWallet(bob)
	handle(err)
	// alicesWallet, _ := wallets.GetWallet(alice)
	genesisConf := blockchain.GenesisConfig{Amount: 50, To: bobsWallet.PublicKey, Sig: nil}

	difficulty := 20
	bch, err := blockchain.NewBlockChain(difficulty, genesisConf, wallets)
	handle(err)

	bobBalance, err := bch.GetBalance(bob)
	handle(err)
	fmt.Printf("Bob's balance = %d\n", bobBalance)

	aliceBalance, err := bch.GetBalance(alice)
	handle(err)
	fmt.Printf("Alice's balance = %d\n", aliceBalance)

	firstTx, err := bch.NewTransaction(bob, alice, 10, wallets)

	handle(err)
	bch.AddBlock([]*blockchain.Transaction{firstTx})

	bobBalance, err = bch.GetBalance(bob)
	handle(err)
	fmt.Printf("Bob's balance = %d\n", bobBalance)
}
