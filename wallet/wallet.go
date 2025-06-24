package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pragmaticcs/blockchain/crypto"
)

type Address string

const (
	AddressLength  = 4
	AddressVersion = 0
	FilePath       = "wallets.json"
)

type Wallets struct {
	Wallets map[Address]*Wallet
}

type Wallet struct {
	PrivateKey    ecdsa.PrivateKey
	PublicKey     crypto.PubKey
	AddressLength int
}

func (w *Wallets) AddWallet() (Address, error) {
	newWallet, err := NewWallet()
	if err != nil {
		return "", err
	}
	address, err := newWallet.GetAddress()
	if err != nil {
		return "", err
	}
	w.Wallets[address] = newWallet
	return address, nil
}

func (w *Wallets) GetWallet(addr Address) (*Wallet, error) {
	wallet, ok := w.Wallets[addr]
	if !ok {
		return nil, NewWalletError(fmt.Sprintf("wallet with address [%s] doesn't exist", addr))
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

func InitWallets() (*Wallets, error) {
	wallets, err := LoadWalletsFromFile()
	if os.IsNotExist(err) {
		// If the file doesn't exist - return empty struct
		wallets = &Wallets{}
		err = nil
	}
	return wallets, err
}

func NewWallet() (*Wallet, error) {
	private, public, err := crypto.NewKeyPair()
	if err != nil {
		return nil, err
	}

	return &Wallet{
		PrivateKey:    private,
		PublicKey:     public,
		AddressLength: AddressLength,
	}, nil
}

func (w *Wallet) GetAddress() (Address, error) {
	hash, err := crypto.HashPublicKey(w.PublicKey)
	if err != nil {
		return "", err
	}
	versionedHash := append([]byte{byte(AddressVersion)}, hash...)
	checksum := crypto.Checksum(versionedHash, w.AddressLength)

	fullHash := append(versionedHash, checksum...)
	address := crypto.EncodeToBase58(fullHash)
	return Address(address), nil
}

func SaveWalletsToFile(w *Wallets) error {
	var b []byte
	buf := bytes.NewBuffer(b)
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(w); err != nil {
		return err
	}
	return os.WriteFile(FilePath, buf.Bytes(), 0644)
}

func LoadWalletsFromFile() (*Wallets, error) {
	if _, err := os.Stat(FilePath); os.IsNotExist(err) {
		return nil, err
	}
	fileBytes, err := os.ReadFile(FilePath)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(fileBytes)
	decoder := json.NewDecoder(buf)
	var wallets Wallets
	decoder.Decode(&wallets)
	return &wallets, nil
}

type WalletError struct {
	Err string
}

func NewWalletError(e string) WalletError {
	return WalletError{Err: e}
}

func (e WalletError) Error() string {
	return fmt.Sprintf("Wallet error: %s", e.Err)
}
