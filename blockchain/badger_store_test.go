package blockchain

import (
	"fmt"
	"testing"

	"github.com/pragmaticcs/blockchain/wallet"
	"github.com/stretchr/testify/assert"
)

func TestBadgerStore(t *testing.T) {
	tempDir := t.TempDir()
	handle := func(err error) {
		if err != nil {
			t.Error(err)
		}
	}
	wallets := &wallet.Wallets{
		Wallets:      map[wallet.Address]*wallet.Wallet{},
		SaveFilePath: "wallets.json",
	}
	bob, err := wallets.AddWallet()
	handle(err)
	bobsWallet, err := wallets.GetWallet(bob)
	handle(err)
	difficulty := 16
	genesisConf := GenesisConfig{Amount: 50, To: bobsWallet.PublicKey, Sig: nil}

	store, err := NewBadgerStore(fmt.Sprintf("%s/badger", tempDir))
	handle(err)
	defer store.Close()
	bch, err := NewBlockchain(difficulty, genesisConf, store)
	handle(err)

	coinbase, err := NewCoinbase(100, bobsWallet.PublicKey)
	handle(err)
	block, err := bch.AddBlock([]*Transaction{coinbase})
	handle(err)

	blockFromDB, err := store.GetBlock(block.Hash)
	handle(err)
	assert.EqualValues(t, block, blockFromDB)
}
