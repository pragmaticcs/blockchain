package main

import (
	"log"
	"os"

	"github.com/pragmaticcs/blockchain/node"
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
	dir, err := os.Getwd()
	handle(err)

	// satoshiWallet, err := wallet.LoadWalletFromFile(dir, "satoshi")
	// handle(err)

	aliceWallet, err := wallet.LoadWalletFromFile(dir, "alice")
	handle(err)

	aliceAddr, _ := aliceWallet.GetAddress()

	firstNodeOpts := node.NodeOpts{
		Port:                  "localhost:3000",
		SeedNodes:             nil,
		WalletName:            "satoshi",
		WalletDir:             dir,
		BlockchainDatabaseDir: dir,
		GenesisConfigDir:      dir,
	}

	secondNodeOpts := node.NodeOpts{
		Port:                  "localhost:3333",
		SeedNodes:             []string{"localhost:3000"},
		WalletName:            "alice",
		WalletDir:             dir,
		BlockchainDatabaseDir: dir,
		GenesisConfigDir:      dir,
	}

	firstNode, err := node.NewNode(firstNodeOpts)
	handle(err)

	// Now the first blockchain is longer then the second
	err = firstNode.StateManager.Send(aliceAddr, 10)
	handle(err)

	secondNode, err := node.NewNode(secondNodeOpts)
	handle(err)

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
