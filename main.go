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
	wallets, err := wallet.InitWallets("wallets.json")
	handle(err)

	bob, err := wallets.AddWallet()
	handle(err)
	alice, err := wallets.AddWallet()
	handle(err)

	bobsWallet, err := wallets.GetWallet(bob)
	handle(err)
	genesisConf := blockchain.GenesisConfig{Amount: 50, To: bobsWallet.PublicKey, Sig: nil}

	difficulty := 22
	bch, err := blockchain.NewBlockChain(difficulty, genesisConf)
	handle(err)
	stateManager := blockchain.NewStateManager(bch, wallets)

	bobBalance, err := stateManager.GetBalance(bob)
	handle(err)
	fmt.Printf("Bob's balance = %d\n", bobBalance)

	aliceBalance, err := stateManager.GetBalance(alice)
	handle(err)
	fmt.Printf("Alice's balance = %d\n", aliceBalance)

	firstTx, err := bch.NewTransaction(bob, alice, 10, wallets)
	handle(err)
	verified, err := bch.VerifyTransaction(firstTx)
	handle(err)
	if verified {
		fmt.Println("Verified transaction")
	}

	handle(err)
	bch.AddBlock([]*blockchain.Transaction{firstTx})
	fmt.Println("Mining new block with transaction")

	bobBalance, err = stateManager.GetBalance(bob)
	handle(err)
	fmt.Printf("Bob's balance = %d\n", bobBalance)

	aliceBalance, err = stateManager.GetBalance(alice)
	handle(err)
	fmt.Printf("Alice's balance = %d\n", aliceBalance)
}
