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
	"github.com/pragmaticcs/blockchain/wallet"
	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

type (
	tomlConfig struct {
		Node tomlNode `toml:"node"`
	}
	tomlNode struct {
		Wallet tomlWallet `toml:"wallet"`
		Server tomlServer `toml:"server"`
	}
	tomlWallet struct {
		Dir  string `toml:"dir"`
		Name string `toml:"name"`
	}
	tomlServer struct {
		Port  string   `toml:"port"`
		Peers []string `toml:"peers"`
	}
)

func HandleStartNodeDaemon(cmd *cobra.Command, args []string) {
	configPath := fmt.Sprintf("%s/gochain.toml", dataDir)

	var config tomlConfig

	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		log.Printf("Failed to parse config file: %v", err)
	}

	nodeOpts := node.NodeOpts{
		Port:                  config.Node.Server.Port,
		SeedNodes:             config.Node.Server.Peers,
		DataDir:               dataDir,
		WalletName:            config.Node.Wallet.Name,
		WalletDir:             fmt.Sprintf("%s/%s", dataDir, config.Node.Wallet.Dir),
		BlockchainDatabaseDir: fmt.Sprintf("%s/db", dataDir),
		GenesisConfigDir:      dataDir,
	}

	log.Println("Starting node daemon...")
	node, err := node.NewNode(nodeOpts)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	err = node.Start()
	if err != nil {
		log.Printf("node exited with error: %s", err)
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
		log.Fatalf("Failed to open log file: %s\n", err)
	}

	command.Stdout = file
	command.Stderr = file

	if err := command.Start(); err != nil {
		log.Fatalf("Failed to start node daemon: %s\n", err)
	}

	fmt.Printf("Node started as background process with PID: %d\n", command.Process.Pid)
}

func HandleNodeStart(cmd *cobra.Command, args []string) {
	if daemonMode {
		HandleStartNodeDaemon(cmd, args)
	} else {
		HandleSpawnNodeDaemon(cmd, args)
	}
}

func HandleNodeStop(cmd *cobra.Command, args []string) {
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

func HandleNodeSend(cmd *cobra.Command, args []string) {
	payload := node.IpcPayloadSend{
		Address: sendAddress,
		Amount:  sendAmount,
		Fee:     sendFee,
	}
	resp, err := SendIpcRequest(node.IpcSend, payload)
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

func HandleWalletGetAddress(cmd *cobra.Command, args []string) {
	wallet, err := wallet.LoadWalletFromFile(walletDir, walletName)
	if err != nil {
		fmt.Printf("error while loading wallet: %s\n", err)
		return
	}
	addr, _ := wallet.GetAddress()
	fmt.Printf("wallet address: %s\n", addr)
}

func HandleWalletNew(cmd *cobra.Command, args []string) {
	wallet, err := wallet.NewWallet()
	if err != nil {
		fmt.Printf("error while creating a new wallet: %s\n", err)
		return
	}
	err = wallet.SaveToFile(walletDir, walletName)
	if err != nil {
		fmt.Printf("error while saving created wallet to file: %s\n", err)
		return
	}
}

func SendIpcRequest(command node.IpcCommand, payload any) (*node.IpcResponse, error) {
	socketPath := fmt.Sprintf("%s/gochain.sock", dataDir)
	conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
	if err != nil {
		log.Printf("error connecting to node daemon at path %s\n", socketPath)
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
