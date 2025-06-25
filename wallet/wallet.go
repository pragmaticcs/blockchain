package wallet

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/pragmaticcs/blockchain/crypto"
)

type Address string

const (
	AddressLength  = 4
	AddressVersion = 0
)

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

type WalletError struct {
	Err string
}

func NewWalletError(e string) WalletError {
	return WalletError{Err: e}
}

func (e WalletError) Error() string {
	return fmt.Sprintf("Wallet error: %s", e.Err)
}
