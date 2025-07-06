package main

import (
	"fmt"
	"log"

	"github.com/pragmaticcs/blockchain/core"
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

	genesisConf := core.GenesisConfig{Amount: 50, To: bobsWallet.PublicKey}

	store, err := core.NewBadgerStore("badger")
	handle(err)
	defer store.Close()

	difficulty := 10
	bch, err := core.NewBlockchain(difficulty, genesisConf, store)
	handle(err)
	stateManager := core.NewStateManager(bch, wallets)

	bobBalance, err := stateManager.GetBalance(bob)
	handle(err)
	fmt.Printf("Bob's balance = %d\n", bobBalance)

	aliceBalance, err := stateManager.GetBalance(alice)
	handle(err)
	fmt.Printf("Alice's balance = %d\n", aliceBalance)

	err = stateManager.Send(bob, alice, 10)
	handle(err)

	bobBalance, err = stateManager.GetBalance(bob)
	handle(err)
	fmt.Printf("Bob's balance = %d\n", bobBalance)

	aliceBalance, err = stateManager.GetBalance(alice)
	handle(err)
	fmt.Printf("Alice's balance = %d\n", aliceBalance)

	err = stateManager.Send(alice, bob, 5)
	handle(err)

	bobBalance, err = stateManager.GetBalance(bob)
	handle(err)
	fmt.Printf("Bob's balance = %d\n", bobBalance)

	aliceBalance, err = stateManager.GetBalance(alice)
	handle(err)
	fmt.Printf("Alice's balance = %d\n", aliceBalance)
}
