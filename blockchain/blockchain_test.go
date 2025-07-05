package blockchain

import (
	"fmt"
	"testing"

	"github.com/pragmaticcs/blockchain/wallet"
	"github.com/stretchr/testify/assert"
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
	genesisConf := GenesisConfig{Amount: 50, To: bobsWallet.PublicKey}

	difficulty := 10
	store, err := NewBadgerStore(fmt.Sprintf("%s/badger", tempDir))
	handle(err)

	bch, err := NewBlockchain(difficulty, genesisConf, store)
	handle(err)

	stateManager := NewStateManager(bch, wallets)

	bobBalance, err := stateManager.GetBalance(bob)
	handle(err)
	t.Logf("Bob's balance = %d\n", bobBalance)
	assert.EqualValues(t, bobBalance, 50)

	aliceBalance, err := stateManager.GetBalance(alice)
	handle(err)
	t.Logf("Alice's balance = %d\n", aliceBalance)
	assert.EqualValues(t, aliceBalance, 0)

	stateManager.Send(bob, alice, 10)

	bobBalance, err = stateManager.GetBalance(bob)
	handle(err)
	t.Logf("Bob's balance = %d\n", bobBalance)
	assert.EqualValues(t, bobBalance, 40)

	aliceBalance, err = stateManager.GetBalance(alice)
	handle(err)
	t.Logf("Alice's balance = %d\n", aliceBalance)
	assert.EqualValues(t, aliceBalance, 10)
}
