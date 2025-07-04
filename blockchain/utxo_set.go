package blockchain

import (
	"bytes"
	"encoding/binary"
	"sync"
)

type OutPoint struct {
	TxID  TransactionID
	OutID int64
}

func NewOutPoint(txID TransactionID, outID int64) OutPoint {
	return OutPoint{
		TxID:  txID,
		OutID: outID,
	}
}

func EncodeOutpoint(o *OutPoint) ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	err := binary.Write(buf, binary.LittleEndian, o.OutID)
	if err != nil {
		return nil, err
	}
	return append(o.TxID[:], buf.Bytes()...), nil
}

func DecodeOutpoint(data []byte) (*OutPoint, error) {
	txID := TransactionID(data[:32])
	var outID int64
	buf := bytes.NewBuffer(data[32:])
	err := binary.Read(buf, binary.LittleEndian, &outID)
	if err != nil {
		return nil, err
	}
	outpoint := NewOutPoint(txID, outID)
	return &outpoint, nil
}

type UTXOs map[OutPoint]TxOutput

type UTXOset struct {
	Blockchain *Blockchain
	Store      Store

	lock sync.RWMutex
}

func NewUTXOset(b *Blockchain, s Store) *UTXOset {
	return &UTXOset{
		Blockchain: b,
		Store:      s,
		lock:       sync.RWMutex{},
	}
}

func (u *UTXOset) Lock() {
	u.lock.Lock()
}

func (u *UTXOset) Unlock() {
	u.lock.Unlock()
}

func (u *UTXOset) Reindex() error {
	utxos := u.FindUtxoSet()
	err := u.Store.DeleteAllUTXOs()
	if err != nil {
		return err
	}
	err = u.Store.PutUTXOs(utxos)
	if err != nil {
		return err
	}
	return nil
}

func (u *UTXOset) Update(block *Block) error {
	newOutputs := make(UTXOs)
	var spentOutputs []OutPoint

	// Iterate over each new transaction
	for _, tx := range block.Transactions {
		// Process Ouputs
		for outID, out := range tx.Outputs {
			outPoint := NewOutPoint(tx.ID, int64(outID))
			newOutputs[outPoint] = out
		}
		// Process inputs
		// Skip if coinbase
		if tx.IsCoinbase() {
			continue
		}
		for _, in := range tx.Inputs {
			outPoint := NewOutPoint(in.TxID, in.OutID)
			spentOutputs = append(spentOutputs, outPoint)
		}
	}
	err := u.Store.DeleteUTXOs(spentOutputs)
	if err != nil {
		return err
	}
	return u.Store.PutUTXOs(newOutputs)
}

func (u *UTXOset) FindUtxos(pubKeyHash []byte) []TxOutput {
	// Returns all the unspent outputs of an account with pubKeyHash
	var utxos []TxOutput
	_ = u.Store.ForEachUTXO(func(outPoint OutPoint, out TxOutput) error {
		if out.IsLockedWithKey(pubKeyHash) {
			utxos = append(utxos, out)
		}
		return nil
	})
	return utxos
}

func (u *UTXOset) FindSpendableAmount(pubKeyHash []byte) (int, error) {
	// Returns the balance of an account with pubKeyHash
	utxos := u.FindUtxos(pubKeyHash)
	var balance int64 = 0
	for _, utxo := range utxos {
		balance += utxo.Value
	}
	return int(balance), nil
}

func (u *UTXOset) FindSpendableOutputs(pubKeyHash []byte, amount int) (int64, []OutPoint) {
	// Returns the spendable outputs of an account with pubKeyHash
	var utxos []OutPoint
	var currSum int64 = 0
	_ = u.Store.ForEachUTXO(func(outPoint OutPoint, out TxOutput) error {
		if out.IsLockedWithKey(pubKeyHash) && currSum < int64(amount) {
			currSum += out.Value
			utxos = append(utxos, outPoint)
			if currSum >= int64(amount) {
				return nil
			}
		}
		return nil
	})
	return currSum, utxos
}

func (u *UTXOset) FindUtxoSet() UTXOs {
	UTXO := make(UTXOs)
	spentTXOS := make(map[OutPoint]struct{})

	for block := range u.Blockchain.Iterate() {
		// For each transaction in a block
		for _, tx := range block.Transactions {
			// INPUTS - add outputs that have already been spent to the STXO set
			// Ignore coinbases
			if !tx.IsCoinbase() {
				// For each transaction input
				for _, in := range tx.Inputs {
					// Add the outpoint to the set of spent outputs
					outPoint := NewOutPoint(in.TxID, in.OutID)
					spentTXOS[outPoint] = struct{}{}
				}
			}
			// OUTPUTS - collect all outputs of the transaction
			// For each output in this transaction
			for outID, txOut := range tx.Outputs {
				outPoint := NewOutPoint(tx.ID, int64(outID))
				// Check if it is spent
				if _, spent := spentTXOS[outPoint]; !spent {
					// If not spent, add to the UTXO set
					UTXO[outPoint] = txOut
				}
			}
		}
	}
	return UTXO
}
