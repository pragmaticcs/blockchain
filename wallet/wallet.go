package wallet

import (
	"crypto/ecdsa"

	"github.com/pragmaticcs/blockchain/crypto"
)

const (
	AddressLength  = 4
	AddressVersion = 0
)

type Wallets struct {
	Wallets map[string]*Wallet
}

type Wallet struct {
	PrivateKey    ecdsa.PrivateKey
	PublicKey     crypto.PubKey
	AddressLength int
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

func (w *Wallet) GetAddress() ([]byte, error) {
	hash, err := crypto.HashPublicKey(w.PublicKey)
	if err != nil {
		return nil, err
	}
	versionedHash := append([]byte{byte(AddressVersion)}, hash...)
	checksum := crypto.Checksum(versionedHash, w.AddressLength)

	fullHash := append(versionedHash, checksum...)
	address := crypto.EncodeToBase58(fullHash)
	return address, nil
}
