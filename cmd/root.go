package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gochain",
	Short: "A simple blockchain implementation in pure Golang",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var daemonMode bool
var dataDir string
var walletDir string
var walletName string
var listenAddr string
var seedPeers []string
var sendAddress string
var sendAmount uint64
var sendFee uint64

func init() {
	// Root command
	rootCmd.AddCommand(nodeCmd, walletCmd)

	// Node command
	nodeCmd.AddCommand(nodeStartCmd, nodeStopCmd, nodeSendCmd)
	nodeCmd.PersistentFlags().StringVar(&dataDir, "dataDir", "./", "Data directory containing genesis.data, database and logs")

	// Node start command
	nodeStartCmd.Flags().BoolVar(&daemonMode, "daemon", false, "Run in daemon mode")
	nodeStartCmd.Flags().MarkHidden("daemon")
	nodeStartCmd.Flags().StringVar(&listenAddr, "listenAddr", ":3000", "Address to listen on")
	nodeStartCmd.Flags().StringVar(&walletName, "walletName", "", "Name of the wallet [name]_wallet.json the node will use")
	nodeStartCmd.Flags().StringVar(&walletDir, "walletDir", "", "Directory in which the wallet is stored")
	nodeStartCmd.Flags().StringSliceVar(&seedPeers, "seedPeers", nil, "Comma-separated list of seed peer addresses to connect to")

	// Node send command
	nodeSendCmd.Flags().Uint64Var(&sendAmount, "amount", 0, "Amount of coins to send to address")
	nodeSendCmd.Flags().Uint64Var(&sendFee, "fee", 0, "Mining fee amount")
	nodeSendCmd.Flags().StringVar(&sendAddress, "address", "", "Address of the reciever")

	// Wallet command
	walletCmd.AddCommand(walletGetAddress, walletNew)
	walletCmd.PersistentFlags().StringVar(&walletName, "name", "my", "Name of the wallet [name]_wallet.json")
	walletCmd.PersistentFlags().StringVar(&walletDir, "dir", "./", "Directory in which the wallet is stored")
}

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage the local blockchain node",
}

var nodeStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the local blockchain node",
	Run:   HandleNodeStart,
}

var nodeStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the local blockchain node",
	Run:   HandleNodeStop,
}

var nodeSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Sends [amount] from nodes wallet to given [address]",
	Run:   HandleNodeSend,
}

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Manage your blockchain wallet",
}

var walletGetAddress = &cobra.Command{
	Use:   "address",
	Short: "Returns address of the wallet",
	Run:   HandleWalletGetAddress,
}

var walletNew = &cobra.Command{
	Use:   "new",
	Short: "Creates a new wallet with [name] in a [directory]",
	Run:   HandleWalletNew,
}
