package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"github.com/pragmaticcs/blockchain/crypto"
)

type TransactionID = [32]byte

var TransactionZeroID TransactionID

type Transaction struct {
	ID      TransactionID
	Inputs  []TxInput
	Outputs []TxOutput
}

func (t *Transaction) GetHash() [32]byte {
	tx := *t
	tx.ID = [32]byte{}
	data := tx.GetData()
	hash := sha256.Sum256(data)
	return hash
}

func (t *Transaction) GetData() []byte {
	var outData []byte
	for _, out := range t.Outputs {
		outData = append(outData, out.GetData()...)
	}
	var inData []byte
	for _, in := range t.Inputs {
		inData = append(inData, in.GetData()...)
	}
	return bytes.Join([][]byte{
		t.ID[:],
		outData,
		inData,
	}, []byte{})
}

type TxOutput struct {
	Value      int64 // Amount of coins
	PubKeyHash []byte
}

func (out *TxOutput) GetData() []byte {
	return bytes.Join([][]byte{
		GetHexValue(out.Value),
		out.PubKeyHash,
	}, []byte{})
}

func NewTxOutput(value int64, address []byte) *TxOutput {
	out := &TxOutput{
		Value:      value,
		PubKeyHash: nil,
	}
	LockOutput(out, address)
	return out
}

type TxInput struct {
	TxID   TransactionID // Points to the ID of the Transaction
	OutID  int64         // Points to the index of Tx in the list of Tx's inside the Transaction [ID]
	Sig    crypto.Signature
	PubKey crypto.PubKey
}

func (in *TxInput) GetData() []byte {
	return bytes.Join([][]byte{
		in.TxID[:],
		GetHexValue(in.OutID),
		in.Sig,
		in.PubKey,
	}, []byte{})
}

func NewCoinbase(value int64, pubKeyHash []byte) *Transaction {
	txOut := TxOutput{
		Value:      value,
		PubKeyHash: pubKeyHash,
	}
	txIn := TxInput{
		TxID:  TransactionZeroID, // Empty ID
		OutID: -1,                // No Reference
		Sig:   nil,
	}
	coinbase := &Transaction{
		ID:      [32]byte{},
		Inputs:  []TxInput{txIn},
		Outputs: []TxOutput{txOut},
	}
	coinbase.ID = coinbase.GetHash()
	return coinbase
}

func (t *Transaction) IsCoinbase() bool {
	// Checks if a transaction is a Coinbase
	return len(t.Inputs) == 1 && t.Inputs[0].TxID == TransactionZeroID && t.Inputs[0].OutID == -1
}

func (out *TxOutput) IsLockedWithKey(pubkeyHash []byte) bool {

	return bytes.Equal(out.PubKeyHash, pubkeyHash)
}

func (in *TxInput) UsesKey(pubkeyHash []byte) (bool, error) {
	lockingHash, err := crypto.HashPublicKey(in.PubKey)
	if err != nil {
		return false, err
	}
	return bytes.Equal(lockingHash, pubkeyHash), nil
}

func LockOutput(out *TxOutput, address []byte) {
	pubkeyHash := crypto.DecodeFromBase58(address)
	pubkeyHash = pubkeyHash[1 : len(pubkeyHash)-4]
	out.PubKeyHash = pubkeyHash
}

func (tx *Transaction) Copy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput
	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{
			TxID:   in.TxID,
			OutID:  in.OutID,
			Sig:    in.Sig,
			PubKey: in.PubKey,
		})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{
			Value:      out.Value,
			PubKeyHash: out.PubKeyHash,
		})
	}
	return Transaction{
		ID:      tx.ID,
		Inputs:  inputs,
		Outputs: outputs,
	}
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTxs map[TransactionID]Transaction) error {
	if tx.IsCoinbase() {
		return NewTransactionError("coinbase can't be signed")
	}

	for _, in := range tx.Inputs {
		// Check if inputs are valid
		if prevTxs[in.TxID].ID == TransactionZeroID {
			return NewTransactionError("previous transaction doesn't exist")
		}
	}
	txCopy := tx.Copy()

	for id, in := range tx.Inputs {
		prevTx := prevTxs[in.TxID]
		txCopy.Inputs[id].Sig = nil
		txCopy.Inputs[id].PubKey = prevTx.Outputs[in.OutID].PubKeyHash

		txID := txCopy.GetHash()

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txID[:])
		if err != nil {
			return err
		}
		sig := crypto.SerializeEllipticCoordinates(r, s)

		tx.Inputs[id].Sig = sig
	}
	return nil
}

func (tx *Transaction) Verify(prevTxs map[TransactionID]Transaction) (bool, error) {
	if tx.IsCoinbase() {
		return true, nil
	}

	for _, in := range tx.Inputs {
		// Check if inputs are valid
		if prevTxs[in.TxID].ID == TransactionZeroID {
			return false, NewTransactionError("previous transaction doesn't exist")
		}
	}
	txCopy := tx.Copy()
	curve := elliptic.P256()

	for id, in := range txCopy.Inputs {
		prevTx := prevTxs[in.TxID]
		txCopy.Inputs[id].Sig = nil
		txCopy.Inputs[id].PubKey = prevTx.Outputs[in.OutID].PubKeyHash
		txID := txCopy.GetHash()

		r, s := crypto.DeserializeEllipticCoordinates(in.Sig)
		x, y := crypto.DeserializeEllipticCoordinates(in.PubKey)

		ecdsaPubKey := ecdsa.PublicKey{Curve: curve, X: x, Y: y}

		if !ecdsa.Verify(&ecdsaPubKey, txID[:], r, s) {
			return false, nil
		}
	}
	return true, nil
}

func (b *Blockchain) FindTransaction(id TransactionID) (*Transaction, error) {
	for block := range b.Iterate() {
		for _, tx := range block.Transactions {
			if tx.ID == id {
				return &tx, nil
			}
		}
	}
	return nil, NewTransactionError("transaction doesn't exist")
}

func (b *Blockchain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) error {
	if tx.IsCoinbase() {
		return NewTransactionError("coinbase can't be signed")
	}
	prevTxs := make(map[TransactionID]Transaction)
	for _, in := range tx.Inputs {
		prevTx, err := b.FindTransaction(in.TxID)
		if err != nil {
			return err
		}
		prevTxs[prevTx.ID] = *prevTx
	}
	return tx.Sign(privateKey, prevTxs)
}

func (b *Blockchain) VerifyTransaction(tx *Transaction) (bool, error) {
	if tx.IsCoinbase() {
		return true, nil
	}
	prevTxs := make(map[TransactionID]Transaction)

	for _, in := range tx.Inputs {
		prevTx, err := b.FindTransaction(in.TxID)
		if err != nil {
			return false, err
		}
		prevTxs[prevTx.ID] = *prevTx
	}
	return tx.Verify(prevTxs)
}

type TransactionError struct {
	Err string
}

func NewTransactionError(e string) TransactionError {
	return TransactionError{Err: e}
}

func (e TransactionError) Error() string {
	return fmt.Sprintf("Transaction error: %s", e.Err)
}
