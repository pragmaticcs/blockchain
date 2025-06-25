package wallet

import (
	"bytes"
	"encoding/json"
	"os"
)

var (
	ErrorWalletNotFound      = NewWalletError("wallet not found")
	ErrorInsufficientBalance = NewWalletError("insufficient wallet balance")
)

type Wallets struct {
	// For storing all wallets (currently for testing purposes)
	Wallets      map[Address]*Wallet
	SaveFilePath string
}

func (w *Wallets) GetWallet(addr Address) (*Wallet, error) {
	wallet, ok := w.Wallets[addr]
	if !ok {
		return nil, ErrorWalletNotFound
	}
	return wallet, nil
}

func (w *Wallets) GetAllAddresses() []Address {
	var addresses []Address

	for a := range w.Wallets {
		addresses = append(addresses, a)
	}
	return addresses
}

func InitWallets(saveFilePath string) (*Wallets, error) {
	wallets, err := LoadWalletsFromFile(saveFilePath)
	if os.IsNotExist(err) {
		// If the file doesn't exist - return empty struct
		wallets = &Wallets{}
		err = nil
	}
	return wallets, err
}

func (w *Wallets) SaveWalletsToFile() error {
	var b []byte
	buf := bytes.NewBuffer(b)
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(w); err != nil {
		return err
	}
	return os.WriteFile(w.SaveFilePath, buf.Bytes(), 0644)
}

func LoadWalletsFromFile(saveFilePath string) (*Wallets, error) {
	if _, err := os.Stat(saveFilePath); os.IsNotExist(err) {
		return nil, err
	}
	fileBytes, err := os.ReadFile(saveFilePath)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(fileBytes)
	decoder := json.NewDecoder(buf)
	var wallets Wallets
	decoder.Decode(&wallets)
	return &wallets, nil
}

func (w *Wallets) AddressExists(addr Address) bool {
	_, exists := w.Wallets[addr]
	return exists
}
