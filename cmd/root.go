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
		// Do Stuff Here
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
var walletName string
var listenAddr string
var seedPeers []string

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.AddCommand(nodeStartCmd, nodeStopCmd)

	nodeStartCmd.Flags().BoolVar(&daemonMode, "daemon", false, "Run in daemon mode")
	nodeStartCmd.Flags().MarkHidden("daemon")
	nodeStartCmd.Flags().StringVar(&listenAddr, "listenAddr", "localhost:3000", "Address to listen on")

	rootCmd.PersistentFlags().StringVar(&dataDir, "dataDir", "./", "Data directory containing wallet.json, genesis.data, database and logs")
	rootCmd.PersistentFlags().StringVar(&walletName, "walletName", "john", "Wallet name (e.g. [walletName]_wallet.json)")
	rootCmd.PersistentFlags().StringSliceVar(&seedPeers, "seedPeers", nil, "Comma-separated list of seed peer addresses to connect to")
}

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage the local blockchain node.",
}

var nodeStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the local blockchain node",
	Run:   HandleStartNode,
}

var nodeStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the local blockchain node",
	Run:   HandleStopNode,
}
