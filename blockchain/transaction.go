package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
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
	Value  int64 // Amount of coins
	PubKey crypto.PubKey
}

type TxInput struct {
	TxID  TransactionID // Points to the ID of the Transaction
	OutID int           // Points to the index of Tx in the list of Tx's inside the Transaction [ID]
	Sig   crypto.Signature
}

func NewCoinbase(value int64, to crypto.PubKey, sig crypto.Signature) (*Transaction, error) {
	txOut := TxOutput{
		Value:  value,
		PubKey: to,
	}
	txIn := TxInput{
		TxID:  nil, // Empty ID
		OutID: -1,  // No Reference
		Sig:   sig,
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
	// Checks if a transaction is a Coinbase
	return len(t.Inputs) == 1 && t.Inputs[0].TxID == nil && t.Inputs[0].OutID == -1
}

func (in *TxInput) CanUnlock(sig crypto.Signature) bool {
	// Currently a temporary function
	return slices.Equal(in.Sig, sig)
}

func (out *TxOutput) CanBeUnlocked(pub crypto.PubKey) bool {
	// Currently a temporary function
	return slices.Equal(out.PubKey, pub)
}

func (b *BlockChain) FindUnspentTransactions(pub crypto.PubKey) []*Transaction {
	var unspentTxs []*Transaction
	// Each transaction has a corresponding array of ids
	// that point to the spent outputs in the transaction's array of outputs
	spentTXOs := make(map[string][]int)

	blockchainIterator := b.ToIterator()
	for {
		block := blockchainIterator.Backward()

		if block == nil {
			break
		}

		// For each transaction in a block
		for _, tx := range block.Transactions {
			// Get the txID in string form
			txID := hex.EncodeToString(tx.ID)

			// OUTPUTS - collect all outputs of the transaction
			// For each output in this transaction
		Outputs:
			for i, txOut := range tx.Outputs {
				// Check if transaction is in the set of [tx <==> spent outputs in tx]
				spentOutsIDs, ok := spentTXOs[txID]
				if ok {
					// If transaction is in the set
					// for each unspent transaction
					for _, spentOut := range spentOutsIDs {
						// Check if index of output in the array of outputs == its own index ???
						if spentOut == i {
							continue Outputs
						}
					}
				}
				// Check if the output is owned by the owner of pubkey
				if txOut.CanBeUnlocked(pub) {
					unspentTxs = append(unspentTxs, tx)
				}
			}
			// INPUTS - remove outputs that have already been spent
			// Ignore coinbases
			if !tx.IsCoinbase() {
				// For each input in the transaction's list of inputs
				for _, in := range tx.Inputs {
					// Check if the input is owned by the owner of pubkey
					if in.CanUnlock(pub) {
						inID := hex.EncodeToString(in.TxID)
						// Add the output the input uses to the set of spent outputs
						spentTXOs[inID] = append(spentTXOs[inID], in.OutID)
					}
				}
			}
		}
	}
	return unspentTxs
}

func (b *BlockChain) FindUTXO(address crypto.PubKey) {

}
