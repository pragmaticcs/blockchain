package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/pragmaticcs/blockchain/crypto"
)

var (
	ErrorInsufficientBalance = NewWalletError("insufficient account balance")
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

type JsonWallet struct {
	PrivateKey    string `json:"private_key"`
	PublicKey     string `json:"public_key"`
	AddressLength int    `json:"address_length"`
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

func (w *Wallet) SaveToFile(dir, name string) error {
	path := fmt.Sprintf("%s/%s_wallet.json", dir, name)
	privateKeyBytes, err := x509.MarshalECPrivateKey(&w.PrivateKey)
	if err != nil {
		return err
	}

	privatePem := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyEncoded := hex.EncodeToString(w.PublicKey)

	jsonWallet := JsonWallet{
		PrivateKey:    string(privatePem),
		PublicKey:     string(publicKeyEncoded),
		AddressLength: w.AddressLength,
	}

	var b []byte
	buf := bytes.NewBuffer(b)
	jsonEncoder := json.NewEncoder(buf)
	jsonEncoder.Encode(jsonWallet)

	return os.WriteFile(path, buf.Bytes(), 0600)
}

func LoadWalletFromFile(dir, name string) (*Wallet, error) {
	path := fmt.Sprintf("%s/%s_wallet.json", dir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(fileBytes)
	decoder := json.NewDecoder(buf)

	var jsonWallet JsonWallet
	decoder.Decode(&jsonWallet)

	block, _ := pem.Decode([]byte(jsonWallet.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing private key")
	}
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, err := hex.DecodeString(jsonWallet.PublicKey)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		PrivateKey:    *privateKey,
		PublicKey:     publicKey,
		AddressLength: jsonWallet.AddressLength,
	}, nil
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
