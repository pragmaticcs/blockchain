package core

import (
	"testing"

	"github.com/pragmaticcs/blockchain/wallet"
	"github.com/stretchr/testify/assert"
)

func TestGobEncoderDecoder(t *testing.T) {
	wallets := &wallet.Wallets{
		Wallets:      map[wallet.Address]*wallet.Wallet{},
		SaveFilePath: "wallets.json",
	}
	bob, err := wallets.AddWallet()
	assert.NoError(t, err)
	bobsWallet, err := wallets.GetWallet(bob)
	assert.NoError(t, err)

	coinbase, err := NewCoinbase(100, bobsWallet.PublicKey)
	assert.NoError(t, err)

	block := NewBlock(nil, []*Transaction{coinbase})

	gobEncoder := GobEncoder[Block]{}
	encoding, err := gobEncoder.Encode(block)
	assert.NoError(t, err)

	decoding, err := gobEncoder.Decode(encoding)
	assert.NoError(t, err)
	assert.EqualValues(t, block, decoding)

}
