package main

import (
	"log"
	"os"
	"time"

	"github.com/pragmaticcs/blockchain/node"
	"github.com/pragmaticcs/blockchain/wallet"
)

func handle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	dir, err := os.Getwd()
	handle(err)

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

	thirdNodeOpts := node.NodeOpts{
		Port:                  "localhost:5555",
		SeedNodes:             []string{"localhost:3333"},
		WalletName:            "bob",
		WalletDir:             dir,
		BlockchainDatabaseDir: dir,
		GenesisConfigDir:      dir,
	}

	firstNode, err := node.NewNode(firstNodeOpts)
	handle(err)

	// Now the first blockchain is longer then the second
	for range 10 {
		err = firstNode.StateManager.Send(aliceAddr, 1)
		handle(err)
	}

	secondNode, err := node.NewNode(secondNodeOpts)
	handle(err)

	thirdNode, err := node.NewNode(thirdNodeOpts)
	handle(err)

	defer firstNode.Stop()
	defer secondNode.Stop()
	defer thirdNode.Stop()

	go func() {
		err := firstNode.Start()
		handle(err)
	}()

	time.Sleep(1000 * time.Millisecond)
	go func() {
		err := secondNode.Start()
		handle(err)
	}()

	time.Sleep(1000 * time.Millisecond)
	go func() {
		err := thirdNode.Start()
		handle(err)
	}()

	select {}
}
