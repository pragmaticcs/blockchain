package wallet

import (
	"fmt"
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

func TestWalletsFile(t *testing.T) {
	temp_dir := t.TempDir()
	t.Log(temp_dir)
	walletSet := make(map[Address]*Wallet)
	for i := 0; i < 10; i++ {
		wallet, err := NewWallet()
		if err != nil {
			t.Error(err)
		}
		address, err := wallet.GetAddress()
		if err != nil {
			t.Error(err)
		}
		walletSet[address] = wallet
	}
	wallets := Wallets{
		Wallets:      walletSet,
		SaveFilePath: fmt.Sprintf("%s/wallets.json", temp_dir),
	}
	err := wallets.SaveToFile()
	if err != nil {
		t.Error(err)
	}
	walletsFromFile, err := LoadWalletsFromFile(wallets.SaveFilePath)
	if err != nil {
		t.Error(err)
	}
	counter := 1
	for k := range walletSet {
		t.Logf("Wallet №%d addr (Not saved) [%s]", counter, k)
		counter++
	}
	counter = 1
	for k := range walletsFromFile.Wallets {
		t.Logf("Wallet №%d addr (From saved) [%s]", counter, k)
		counter++
	}
}
