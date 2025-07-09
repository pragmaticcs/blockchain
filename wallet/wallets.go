package wallet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

var (
	ErrorWalletNotFound      = NewWalletError("wallet not found")
	ErrorInsufficientBalance = NewWalletError("insufficient wallet balance")
)

type Wallets struct {
	// For storing wallets of a node
	Wallets map[Address]*Wallet
	FileDir string
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

func InitWallets(FileDir string) (*Wallets, error) {
	wallets, err := LoadWalletsFromFile(FileDir)
	if wallets == nil {
		// If the file doesn't exist - return empty struct
		wallets = &Wallets{
			Wallets: map[Address]*Wallet{},
			FileDir: FileDir,
		}
	}
	return wallets, err
}

func (w *Wallets) getFilePath() string {
	return fmt.Sprintf("%s/wallets.json", w.FileDir)
}

func (w *Wallets) SaveToFile() error {
	var b []byte
	buf := bytes.NewBuffer(b)
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(w); err != nil {
		return err
	}
	return os.WriteFile(w.getFilePath(), buf.Bytes(), 0644)
}

func LoadWalletsFromFile(FileDir string) (*Wallets, error) {
	if _, err := os.Stat(FileDir); os.IsNotExist(err) {
		return nil, nil
	}
	filePath := fmt.Sprintf("%s/wallets.json", FileDir)
	fileBytes, err := os.ReadFile(filePath)
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
