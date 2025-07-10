package main

import (
	"log"

	"github.com/pragmaticcs/blockchain/core"
	"github.com/pragmaticcs/blockchain/wallet"
)

func handle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// func main() {
// 	wallets, err := wallet.InitWallets("wallets.json")
// 	handle(err)

// 	bob, err := wallets.AddWallet()
// 	handle(err)
// 	alice, err := wallets.AddWallet()
// 	handle(err)

// 	bobsWallet, err := wallets.GetWallet(bob)
// 	handle(err)

// 	genesisConf := core.GenesisConfig{Amount: 50, To: bobsWallet.PublicKey}

// 	store, err := core.NewBadgerStore("badger")
// 	handle(err)
// 	defer store.Close()

// 	difficulty := 20
// 	bch, err := core.NewBlockchain(difficulty, genesisConf, store)
// 	handle(err)
// 	stateManager := core.NewStateManager(bch, wallets)

// 	bobBalance, err := stateManager.GetBalance(bob)
// 	handle(err)
// 	fmt.Printf("Bob's balance = %d\n", bobBalance)

// 	aliceBalance, err := stateManager.GetBalance(alice)
// 	handle(err)
// 	fmt.Printf("Alice's balance = %d\n", aliceBalance)

// 	err = stateManager.Send(bob, alice, 10)
// 	handle(err)

// 	bobBalance, err = stateManager.GetBalance(bob)
// 	handle(err)
// 	fmt.Printf("Bob's balance = %d\n", bobBalance)

// 	aliceBalance, err = stateManager.GetBalance(alice)
// 	handle(err)
// 	fmt.Printf("Alice's balance = %d\n", aliceBalance)

// 	err = stateManager.Send(alice, bob, 5)
// 	handle(err)

// 	bobBalance, err = stateManager.GetBalance(bob)
// 	handle(err)
// 	fmt.Printf("Bob's balance = %d\n", bobBalance)

// 	aliceBalance, err = stateManager.GetBalance(alice)
// 	handle(err)
// 	fmt.Printf("Alice's balance = %d\n", aliceBalance)
// }

func main() {
	// Create wallets
	wallets, err := wallet.InitWallets("")
	handle(err)

	bob, err := wallets.AddWallet()
	handle(err)
	_, err = wallets.AddWallet()
	handle(err)

	bobsWallet, err := wallets.GetWallet(bob)
	handle(err)

	genesisConf := core.GenesisConfig{Amount: 50, To: bobsWallet.PublicKey}

	// Create dbs
	firstStore, err := core.NewBadgerStore("")
	handle(err)
	secondStore, err := core.NewBadgerStore("")
	handle(err)

	firstBlockchain, err := core.NewBlockchain(20, genesisConf, firstStore)
	handle(err)

	secondBlockchain, err := core.NewBlockchain(20, genesisConf, secondStore)
	handle(err)

	firstNodeOpts := core.NodeOpts{
		Port:                  "localhost:3000",
		SeedNodes:             nil,
		WalletFileDir:         "",
		BlockchainDatabaseDir: "",
	}

	secondNodeOpts := core.NodeOpts{
		Port:                  "localhost:3333",
		SeedNodes:             []string{"localhost:3000"},
		WalletFileDir:         "",
		BlockchainDatabaseDir: "",
	}

	firstNode := core.NewNodeTesting(firstNodeOpts, firstBlockchain, wallets)
	secondNode := core.NewNodeTesting(secondNodeOpts, secondBlockchain, wallets)
	defer firstNode.Stop()
	defer secondNode.Stop()

	go func() {
		err := firstNode.Start()
		handle(err)
	}()
	go func() {
		err := secondNode.Start()
		handle(err)
	}()

	select {}
}
