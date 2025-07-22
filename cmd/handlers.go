package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

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
	resp, err := SendIpcRequest(node.IpcStopNode, nil)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("node ipc server returned an error: %s\n", resp.Error)
		return
	}

	fmt.Printf("response from node: %v\n", resp.Data)
}

func SendIpcRequest(command node.IpcCommand, payload any) (*node.IpcResponse, error) {
	socketPath := fmt.Sprintf("%s/gochain.sock", dataDir)
	conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
	if err != nil {
		log.Printf("error connecting to node daemon at path %s", socketPath)
	}
	defer conn.Close()

	var rawPayload json.RawMessage
	if payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload into json: %s", err)
		}
		rawPayload = payloadBytes
	}
	request := node.IpcRequest{
		Command: command,
		Payload: rawPayload,
	}
	err = json.NewEncoder(conn).Encode(&request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to daemon: %s", err)
	}

	var response node.IpcResponse
	err = json.NewDecoder(conn).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode daemon response: %s", err)
	}
	return &response, nil
}
