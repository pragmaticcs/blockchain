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
	dir, _ := os.Getwd()

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
		SeedNodes:             []string{"localhost:3333", "localhost:3000"},
		WalletName:            "bob",
		WalletDir:             dir,
		BlockchainDatabaseDir: dir,
		GenesisConfigDir:      dir,
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

	log.Println("sleeping 3 seconds")
	time.Sleep(5 * time.Second)

	log.Println("EXECUTING TRANSACTION")
	err = firstNode.Send(aliceAddr, 10, 3)
	handle(err)

	select {}
}
