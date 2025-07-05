package core

import (
	"testing"

	"github.com/pragmaticcs/blockchain/wallet"
	"github.com/stretchr/testify/assert"
)

func TestGobEncoderDecoder(t *testing.T) {
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

	coinbase, err := NewCoinbase(100, bobsWallet.PublicKey)
	handle(err)

	block := NewBlock(nil, []*Transaction{coinbase})

	gobEncoder := GobEncoder[Block]{}
	encoding, err := gobEncoder.Encode(block)
	handle(err)

	decoding, err := gobEncoder.Decode(encoding)
	handle(err)
	assert.EqualValues(t, block, decoding)

}
