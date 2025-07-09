package core

import (
	"testing"

	"github.com/pragmaticcs/blockchain/wallet"
	"github.com/stretchr/testify/assert"
)

func TestGobEncoderDecoder(t *testing.T) {
	wallets := &wallet.Wallets{
		Wallets: map[wallet.Address]*wallet.Wallet{},
		FileDir: "",
	}
	bob, err := wallets.AddWallet()
	assert.NoError(t, err)
	bobsWallet, err := wallets.GetWallet(bob)
	assert.NoError(t, err)

	coinbase, err := NewCoinbase(100, bobsWallet.PublicKey)
	assert.NoError(t, err)

	block, err := NewBlock(nil, []*Transaction{coinbase}, 2)
	assert.NoError(t, err)

	gobEncoder := GobEncoder[Block]{}
	encoding, err := gobEncoder.Encode(block)
	assert.NoError(t, err)

	decoding, err := gobEncoder.Decode(encoding)
	assert.NoError(t, err)
	assert.EqualValues(t, block, decoding)

}
