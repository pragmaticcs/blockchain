package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"os"

	"github.com/pragmaticcs/blockchain/crypto"
)

const (
	AddressLength  = 4
	AddressVersion = 0
	FilePath       = "wallets.json"
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
