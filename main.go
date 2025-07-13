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

	time.Sleep(2 * time.Second)
	go func() {
		err := secondNode.Start()
		handle(err)
	}()

	select {}
}
