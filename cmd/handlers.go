package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/pragmaticcs/blockchain/node"
	"github.com/spf13/cobra"
)

func HandleStartNodeDaemon(cmd *cobra.Command, args []string) {
	log.Println("Starting node daemon...")

	nodeOpts := node.NodeOpts{
		Port:                  listenAddr,
		SeedNodes:             seedPeers,
		WalletName:            walletName,
		WalletDir:             dataDir,
		BlockchainDatabaseDir: fmt.Sprintf("%s/db", dataDir),
		GenesisConfigDir:      dataDir,
	}

	node, err := node.NewNode(nodeOpts)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	defer node.Stop()
	err = node.Start()
	if err != nil {
		log.Fatalf("Node failed to start: %v", err)
	}
	select {}
}

func HandleSpawnNodeDaemon(cmd *cobra.Command, args []string) {
	daemonArgs := os.Args[1:]
	daemonArgs = append(daemonArgs, "--daemon")

	command := exec.Command(os.Args[0], daemonArgs...)

	logFile := fmt.Sprintf("%s/gochaind.log", dataDir)

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	command.Stdout = file
	command.Stderr = file

	if err := command.Start(); err != nil {
		log.Fatalf("Failed to start node daemon: %v", err)
	}

	fmt.Printf("Node started as background process with PID: %d\n", command.Process.Pid)
}

func HandleStartNode(cmd *cobra.Command, args []string) {
	if daemonMode {
		HandleStartNodeDaemon(cmd, args)
	} else {
		HandleSpawnNodeDaemon(cmd, args)
	}
}

func HandleStopNode(cmd *cobra.Command, args []string) {
	log.Println("Stopping node daemon...")
}
