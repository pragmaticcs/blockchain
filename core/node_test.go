package core

import (
	"testing"
	"time"

	"github.com/pragmaticcs/blockchain/wallet"
	"github.com/stretchr/testify/assert"
)

func TestNode(t *testing.T) {
	tempDir := t.TempDir()
	// Create wallets
	wallets, err := wallet.InitWallets(tempDir)
	assert.NoError(t, err)

	bob, err := wallets.AddWallet()
	assert.NoError(t, err)
	_, err = wallets.AddWallet()
	assert.NoError(t, err)

	bobsWallet, err := wallets.GetWallet(bob)
	assert.NoError(t, err)

	genesisConf := GenesisConfig{Amount: 50, To: bobsWallet.PublicKey}

	// Create dbs
	firstStore, err := NewBadgerStore("")
	assert.NoError(t, err)
	secondStore, err := NewBadgerStore("")
	assert.NoError(t, err)

	firstBlockchain, err := NewBlockchain(defaultDifficulty, genesisConf, firstStore)
	assert.NoError(t, err)

	secondBlockchain, err := NewBlockchain(defaultDifficulty, genesisConf, secondStore)
	assert.NoError(t, err)

	firstNodeOpts := NodeOpts{
		Port:                  "3000",
		SeedNodes:             nil,
		WalletFileDir:         tempDir,
		BlockchainDatabaseDir: tempDir,
	}

	secondNodeOpts := NodeOpts{
		Port:                  "3333",
		SeedNodes:             []string{"3000"},
		WalletFileDir:         tempDir,
		BlockchainDatabaseDir: tempDir,
	}

	firstNode := NewNodeTesting(firstNodeOpts, firstBlockchain, wallets)
	secondNode := NewNodeTesting(secondNodeOpts, secondBlockchain, wallets)

	go firstNode.Start()
	time.Sleep(time.Second * 2)
	secondNode.Start()
}
