package blockchain

import (
	"testing"

	"github.com/pragmaticcs/blockchain/wallet"
)

func TestBlockChain(t *testing.T) {
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
	alice, err := wallets.AddWallet()
	handle(err)

	bobsWallet, err := wallets.GetWallet(bob)
	handle(err)
	genesisConf := GenesisConfig{Amount: 50, To: bobsWallet.PublicKey, Sig: nil}

	difficulty := 20
	bch, err := NewBlockChain(difficulty, genesisConf, wallets)
	handle(err)

	bobBalance, err := bch.GetBalance(bob)
	handle(err)
	t.Logf("Bob's balance = %d\n", bobBalance)

	aliceBalance, err := bch.GetBalance(alice)
	handle(err)
	t.Logf("Alice's balance = %d\n", aliceBalance)

	firstTx, err := bch.NewTransaction(bob, alice, 10, wallets)

	handle(err)
	bch.AddBlock([]*Transaction{firstTx})
	t.Log("Mining new block with transaction\n")

	bobBalance, err = bch.GetBalance(bob)
	handle(err)
	t.Logf("Bob's balance = %d\n", bobBalance)

	aliceBalance, err = bch.GetBalance(alice)
	handle(err)
	t.Logf("Alice's balance = %d\n", aliceBalance)
}
