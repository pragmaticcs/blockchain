package wallet

import (
	"testing"
)

func TestWalletAddress(t *testing.T) {
	wallet, err := NewWallet()
	if err != nil {
		t.Error(err)
	}
	address, err := wallet.GetAddress()
	if err != nil {
		t.Error(err)
	}
	t.Logf("\nAddress: %s\n", string(address))
}
