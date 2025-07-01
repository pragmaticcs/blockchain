package blockchain

import (
	"fmt"
	"testing"

	"github.com/pragmaticcs/blockchain/wallet"
	"github.com/stretchr/testify/assert"
)

func TestGobEncoderDecoder(t *testing.T) {
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

	store, err := NewBadgerStore(fmt.Sprintf("%s/badger", tempDir), &GobBlockEncoder{})
	handle(err)
	bch, err := NewBlockchain(difficulty, genesisConf, store)
	handle(err)

	coinbase, err := NewCoinbase(100, bobsWallet.PublicKey)
	handle(err)
	block, err := bch.AddBlock([]*Transaction{coinbase})
	handle(err)

	gobEncoder := GobBlockEncoder{}
	encoding, err := gobEncoder.Encode(block)
	handle(err)

	decoding, err := gobEncoder.Decode(encoding)
	handle(err)
	assert.EqualValues(t, block, decoding)
}
