package core

import (
	"fmt"
	"testing"

	"github.com/pragmaticcs/blockchain/wallet"
	"github.com/stretchr/testify/assert"
)

func TestBadgerStore(t *testing.T) {
	tempDir := t.TempDir()
	wallets := &wallet.Wallets{
		Wallets: map[wallet.Address]*wallet.Wallet{},
		FileDir: "",
	}
	bob, err := wallets.AddWallet()
	assert.NoError(t, err)
	bobsWallet, err := wallets.GetWallet(bob)
	assert.NoError(t, err)
	difficulty := 16
	genesisConf := GenesisConfig{Amount: 50, To: bobsWallet.PublicKey}

	store, err := NewBadgerStore(fmt.Sprintf("%s/badger", tempDir))

	assert.NoError(t, err)
	defer store.Close()

	bch, err := NewBlockchain(difficulty, genesisConf, store)
	assert.NoError(t, err)

	coinbase, err := NewCoinbase(100, bobsWallet.PublicKey)
	assert.NoError(t, err)
	block, err := bch.NewBlock([]*Transaction{coinbase})
	assert.NoError(t, err)
	err = bch.AddBlock(block)
	assert.NoError(t, err)

	blockFromDB, err := store.GetBlock(block.GetHash())
	assert.NoError(t, err)
	assert.EqualValues(t, block, blockFromDB)
}
