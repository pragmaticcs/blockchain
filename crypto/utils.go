package crypto

import (
	"crypto/sha256"

	"github.com/btcsuite/btcd/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

func HashPublicKey(key PubKey) ([]byte, error) {
	sha256hash := sha256.Sum256(key)
	ripemd := ripemd160.New()
	_, err := ripemd.Write(sha256hash[:])
	if err != nil {
		return nil, err
	}
	publicKeyHash := ripemd.Sum(nil)
	return publicKeyHash, nil
}

func Checksum(data []byte, length int) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:][:length]
}

func EncodeToBase58(data []byte) []byte {
	return []byte(base58.Encode(data))
}

func DecodeFromBase58(data []byte) []byte {
	return base58.Decode(string(data))
}
