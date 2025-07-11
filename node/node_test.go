package node

// func TestNode(t *testing.T) {
// 	tempDir := t.TempDir()
// 	// Create wallets
// 	wallets, err := wallet.InitWallets(tempDir)
// 	assert.NoError(t, err)

// 	bob, err := wallets.AddWallet()
// 	assert.NoError(t, err)
// 	_, err = wallets.AddWallet()
// 	assert.NoError(t, err)

// 	bobsWallet, err := wallets.GetWallet(bob)
// 	assert.NoError(t, err)
// 	pubKeyHash, err := crypto.HashPublicKey(bobsWallet.PublicKey)
// 	assert.NoError(t, err)

// 	genesisConf := core.GenesisConfig{Amount: 50, PubKeyHash: pubKeyHash}

// 	// Create dbs
// 	firstStore, err := core.NewBadgerStore("")
// 	assert.NoError(t, err)
// 	secondStore, err := core.NewBadgerStore("")
// 	assert.NoError(t, err)

// 	firstBlockchain, err := core.NewBlockchain(genesisConf, firstStore)
// 	assert.NoError(t, err)

// 	secondBlockchain, err := core.NewBlockchain(genesisConf, secondStore)
// 	assert.NoError(t, err)

// 	firstNodeOpts := NodeOpts{
// 		Port:                  "3000",
// 		SeedNodes:             nil,
// 		WalletFileDir:         tempDir,
// 		BlockchainDatabaseDir: tempDir,
// 	}

// 	secondNodeOpts := NodeOpts{
// 		Port:                  "3333",
// 		SeedNodes:             []string{"3000"},
// 		WalletFileDir:         tempDir,
// 		BlockchainDatabaseDir: tempDir,
// 	}

// 	firstNode := NewNodeTesting(firstNodeOpts, firstBlockchain, wallets)
// 	secondNode := NewNodeTesting(secondNodeOpts, secondBlockchain, wallets)

// 	go firstNode.Start()
// 	time.Sleep(time.Second * 2)
// 	secondNode.Start()
// }
