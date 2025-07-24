package main

import "github.com/pragmaticcs/blockchain/cmd"

func main() {
	// satoshiWallet, err := wallet.LoadWalletFromFile("data/satoshi", "satoshi")
	// if err != nil {
	// 	log.Panicln(err)
	// }
	// pubKeyHash, _ := crypto.HashPublicKey(satoshiWallet.PublicKey)
	// geneis := core.GenesisConfig{
	// 	Amount:     500,
	// 	PubKeyHash: pubKeyHash,
	// 	Time:       time.Now().Unix(),
	// }
	// geneis.ToFile("./")
	cmd.Execute()
}
