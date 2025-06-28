package crypto

import (
	"bytes"
	"crypto/sha256"
	"math/big"

	"github.com/btcsuite/btcd/btcutil/base58"
	"golang.org/x/crypto/ripemd160" // Bitcoin uses ripemd160 so we going to ignore the deprecated warning
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

func SerializeEllipticCoordinates(x, y *big.Int) []byte {
	xBytes := x.Bytes()
	yBytes := y.Bytes()

	var b []byte
	buf := bytes.NewBuffer(b)
	buf.WriteByte(byte(len(xBytes)))
	buf.Write(xBytes)
	buf.Write(yBytes)
	return buf.Bytes()
}

func DeserializeEllipticCoordinates(sig []byte) (*big.Int, *big.Int) {
	xLen := int(sig[0])
	xBytes := sig[1 : 1+xLen]
	yBytes := sig[1+xLen:]

	x := big.NewInt(0)
	y := big.NewInt(0)
	x = x.SetBytes(xBytes)
	y = y.SetBytes(yBytes)
	return x, y
}
