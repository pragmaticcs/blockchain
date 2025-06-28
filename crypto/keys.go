package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

type PubKey = []byte // 32

type Signature = []byte // 32

func NewKeyPair() (ecdsa.PrivateKey, PubKey, error) {
	// Generate a private/public key pair using the secp256k1 elliptic curve
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return ecdsa.PrivateKey{}, nil, err
	}
	x := privateKey.X
	y := privateKey.Y
	pubKey := SerializeEllipticCoordinates(x, y)
	return *privateKey, pubKey, nil
}
