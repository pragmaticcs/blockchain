package main

import (
	"log"
	"time"

	"github.com/pragmaticcs/blockchain/node"
)

func handle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// bobWallet, err := wallet.LoadWalletFromFile("data/bob", "bob")
	// handle(err)

	// bobAddr, _ := bobWallet.GetAddress()

	// aliceWallet, err := wallet.LoadWalletFromFile("data/alice", "alice")
	// handle(err)

	// aliceAddr, _ := aliceWallet.GetAddress()

	firstNodeOpts := node.NodeOpts{
		Port:                  "localhost:3000",
		SeedNodes:             nil,
		WalletName:            "satoshi",
		WalletDir:             "data/satoshi",
		BlockchainDatabaseDir: "data/satoshi/badger",
		GenesisConfigDir:      "data/satoshi",
	}

	secondNodeOpts := node.NodeOpts{
		Port:                  "localhost:3333",
		SeedNodes:             []string{"localhost:3000"},
		WalletName:            "alice",
		WalletDir:             "data/alice",
		BlockchainDatabaseDir: "data/alice/badger",
		GenesisConfigDir:      "data/alice",
	}

	thirdNodeOpts := node.NodeOpts{
		Port:                  "localhost:5555",
		SeedNodes:             []string{"localhost:3333", "localhost:3000"},
		WalletName:            "bob",
		WalletDir:             "data/bob",
		BlockchainDatabaseDir: "data/bob/badger",
		GenesisConfigDir:      "data/bob",
	}

	firstNode, err := node.NewNode(firstNodeOpts)
	handle(err)

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

	time.Sleep(1 * time.Second)

	go func() {
		err := secondNode.Start()
		handle(err)
	}()

	time.Sleep(1 * time.Second)

	go func() {
		err := thirdNode.Start()
		handle(err)
	}()

	// log.Println("sleeping 3 seconds")
	// time.Sleep(3 * time.Second)

	// log.Println("EXECUTING TRANSACTION")
	// err = secondNode.Send(bobAddr, 5, 1)
	// handle(err)

	select {}
}
