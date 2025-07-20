package core

import (
	"fmt"
	"testing"

	"github.com/pragmaticcs/blockchain/crypto"
	"github.com/pragmaticcs/blockchain/wallet"
	"github.com/stretchr/testify/assert"
)

func TestBadgerStore(t *testing.T) {
	tempDir := t.TempDir()
	bob, err := wallet.NewWallet()
	assert.NoError(t, err)

	pubKeyHash, err := crypto.HashPublicKey(bob.PublicKey)
	assert.NoError(t, err)
	genesisConf := GenesisConfig{Amount: 50, PubKeyHash: pubKeyHash}

	store, err := NewBadgerStore(fmt.Sprintf("%s/badger", tempDir))

	assert.NoError(t, err)
	defer store.Close()

	bch, err := NewBlockchain(genesisConf, store)
	assert.NoError(t, err)

	coinbase := NewCoinbase(100, bob.PublicKey)
	block, err := bch.NewBlock([]Transaction{*coinbase})
	assert.NoError(t, err)
	err = bch.AddBlock(block)
	assert.NoError(t, err)

	blockFromDB, err := store.GetBlock(block.GetHash())
	assert.NoError(t, err)
	assert.EqualValues(t, block, blockFromDB)
}
