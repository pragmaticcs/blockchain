package blockchain

import "github.com/pragmaticcs/blockchain/crypto"

type Transaction struct {
	ID      [32]byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxOutput struct {
	Value  uint64 // Amout of coins
	PubKey crypto.PubKey
}

type TxInput struct {
	ID  [32]byte
	Out int64
	Sig crypto.Signature
}
