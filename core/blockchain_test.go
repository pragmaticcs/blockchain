package core

// func TestBlockChain(t *testing.T) {
// 	tempDir := t.TempDir()
// 	bob, err := wallet.NewWallet()
// 	assert.NoError(t, err)
// 	alice, err := wallet.NewWallet()
// 	assert.NoError(t, err)

// 	assert.NoError(t, err)
// 	pubKeyHash, err := crypto.HashPublicKey(bob.PublicKey)
// 	assert.NoError(t, err)
// 	genesisConf := GenesisConfig{Amount: 50, PubKeyHash: pubKeyHash}

// 	store, err := NewBadgerStore(fmt.Sprintf("%s/badger", tempDir))
// 	assert.NoError(t, err)

// 	bch, err := NewBlockchain(genesisConf, store)
// 	assert.NoError(t, err)

// 	stateManager := NewStateManager(bch, bob)

// 	bobBalance, err := stateManager.GetBalance(bob)
// 	assert.NoError(t, err)
// 	t.Logf("Bob's balance = %d\n", bobBalance)
// 	assert.EqualValues(t, bobBalance, 50)

// 	aliceBalance, err := stateManager.GetBalance(alice)
// 	assert.NoError(t, err)
// 	t.Logf("Alice's balance = %d\n", aliceBalance)
// 	assert.EqualValues(t, aliceBalance, 0)

// 	stateManager.Send(bob, alice, 10)

// 	bobBalance, err = stateManager.GetBalance(bob)
// 	assert.NoError(t, err)
// 	t.Logf("Bob's balance = %d\n", bobBalance)
// 	assert.EqualValues(t, bobBalance, 40)

// 	aliceBalance, err = stateManager.GetBalance(alice)
// 	assert.NoError(t, err)
// 	t.Logf("Alice's balance = %d\n", aliceBalance)
// 	assert.EqualValues(t, aliceBalance, 10)
// }
