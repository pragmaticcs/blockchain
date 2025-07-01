package blockchain

import (
	"fmt"
	"testing"

	"github.com/pragmaticcs/blockchain/wallet"
)

func TestBlockChain(t *testing.T) {
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
	alice, err := wallets.AddWallet()
	handle(err)

	bobsWallet, err := wallets.GetWallet(bob)
	handle(err)
	genesisConf := GenesisConfig{Amount: 50, To: bobsWallet.PublicKey, Sig: nil}

	difficulty := 20
	store, err := NewBadgerStore(fmt.Sprintf("%s/badger", tempDir), &GobBlockEncoder{})
	handle(err)

	bch, err := NewBlockchain(difficulty, genesisConf, store)
	handle(err)

	stateManager := NewStateManager(bch, wallets)

	bobBalance, err := stateManager.GetBalance(bob)
	handle(err)
	t.Logf("Bob's balance = %d\n", bobBalance)

	aliceBalance, err := stateManager.GetBalance(alice)
	handle(err)
	t.Logf("Alice's balance = %d\n", aliceBalance)

	firstTx, err := bch.NewTransaction(bob, alice, 10, wallets)

	handle(err)
	bch.AddBlock([]*Transaction{firstTx})
	t.Log("Mining new block with transaction\n")

	bobBalance, err = stateManager.GetBalance(bob)
	handle(err)
	t.Logf("Bob's balance = %d\n", bobBalance)

	aliceBalance, err = stateManager.GetBalance(alice)
	handle(err)
	t.Logf("Alice's balance = %d\n", aliceBalance)
}
