package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"slices"

	"github.com/pragmaticcs/blockchain/crypto"
)

type TransactionID = []byte

type Transaction struct {
	ID      TransactionID
	Inputs  []TxInput
	Outputs []TxOutput
}

func (t *Transaction) CalculateID() (TransactionID, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(t)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(buf.Bytes())
	return hash[:], nil
}

type TxOutput struct {
	Value  uint64 // Amout of coins
	PubKey crypto.PubKey
}

type TxInput struct {
	ID  TransactionID // Points to the ID of the Transaction
	Out int64         // Points to the index of Tx in the list of Tx's inside the Transaction [ID]
	Sig crypto.Signature
}

func NewCoinbase(value uint64, to crypto.PubKey, sig crypto.Signature) (*Transaction, error) {
	txOut := TxOutput{
		Value:  value,
		PubKey: to,
	}
	txIn := TxInput{
		ID:  []byte{}, // Empty ID
		Out: -1,       // No Reference
		Sig: sig,
	}
	coinbase := &Transaction{
		ID:      nil,
		Inputs:  []TxInput{txIn},
		Outputs: []TxOutput{txOut},
	}
	id, err := coinbase.CalculateID()
	if err != nil {
		return nil, err
	}
	coinbase.ID = id
	return coinbase, nil
}

func (t *Transaction) IsCoinbase() bool {
	return len(t.Inputs) == 1 && len(t.Inputs[0].ID) == 0 && t.Inputs[0].Out == -1
}

func (in *TxOutput) CanUnlock(pub crypto.PubKey) bool {
	return slices.Equal(in.PubKey, pub)
}

func (out *TxInput) CanBeUnlocked(sig crypto.Signature) bool {
	return slices.Equal(out.Sig, sig)
}
