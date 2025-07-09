package core

import (
	"fmt"
	"testing"

	"github.com/pragmaticcs/blockchain/wallet"
	"github.com/stretchr/testify/assert"
)

func TestBlockChain(t *testing.T) {
	tempDir := t.TempDir()
	wallets := &wallet.Wallets{
		Wallets: map[wallet.Address]*wallet.Wallet{},
		FileDir: "",
	}
	bob, err := wallets.AddWallet()
	assert.NoError(t, err)
	alice, err := wallets.AddWallet()
	assert.NoError(t, err)

	bobsWallet, err := wallets.GetWallet(bob)
	assert.NoError(t, err)
	genesisConf := GenesisConfig{Amount: 50, To: bobsWallet.PublicKey}

	difficulty := 10
	store, err := NewBadgerStore(fmt.Sprintf("%s/badger", tempDir))
	assert.NoError(t, err)

	bch, err := NewBlockchain(difficulty, genesisConf, store)
	assert.NoError(t, err)

	stateManager := NewStateManager(bch, wallets)

	bobBalance, err := stateManager.GetBalance(bob)
	assert.NoError(t, err)
	t.Logf("Bob's balance = %d\n", bobBalance)
	assert.EqualValues(t, bobBalance, 50)

	aliceBalance, err := stateManager.GetBalance(alice)
	assert.NoError(t, err)
	t.Logf("Alice's balance = %d\n", aliceBalance)
	assert.EqualValues(t, aliceBalance, 0)

	stateManager.Send(bob, alice, 10)

	bobBalance, err = stateManager.GetBalance(bob)
	assert.NoError(t, err)
	t.Logf("Bob's balance = %d\n", bobBalance)
	assert.EqualValues(t, bobBalance, 40)

	aliceBalance, err = stateManager.GetBalance(alice)
	assert.NoError(t, err)
	t.Logf("Alice's balance = %d\n", aliceBalance)
	assert.EqualValues(t, aliceBalance, 10)
}
