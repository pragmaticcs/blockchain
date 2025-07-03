package blockchain

import (
	"sync"
)

type OutPoint struct {
	TxID  TransactionID
	OutID int
}

type TxOutputs struct {
	Outputs []TxOutput
}

func NewTxOutputs(outs []TxOutput) TxOutputs {
	return TxOutputs{
		Outputs: outs,
	}
}

type UTXOs map[TransactionID]TxOutputs

type UTXOset struct {
	Blockchain *Blockchain
	Store      Store

	lock sync.RWMutex
}

func (u *UTXOset) Lock() {
	u.lock.Lock()
}

func (u *UTXOset) Unlock() {
	u.lock.Unlock()
}

func (u *UTXOset) Reindex() error {
	u.Lock()
	defer u.Unlock()
	utxos := u.Blockchain.FindUtxoSet()
	u.Store.DeleteAllUTXOs()
	u.Store.PutUTXOs(utxos)
	return nil
}

func (u *UTXOset) Update(block *Block) error {
	u.Lock()
	defer u.Unlock()

	stagedOutputs := make(UTXOs)
	// Iterate over each new transaction
	for _, tx := range block.Transactions {
		// Process Ouputs
		stagedOutputs[tx.ID] = NewTxOutputs(tx.Outputs)

		// Process inputs

		// Skip if coinbase
		if tx.IsCoinbase() {
			continue
		}
		for _, in := range tx.Inputs {
			currOutputs, ok := stagedOutputs[in.TxID]
			if !ok {
				fetchedOutputs, err := u.Store.GetUTXO(in.TxID)
				if err != nil {
					return err
				}
				currOutputs = *fetchedOutputs
			}

			for outID, out := range currOutputs.Outputs {
				if outID != in.OutID {
					currOutputs.Outputs = append(currOutputs.Outputs, out)
				}
			}
		}
	}
	return u.Store.PutUTXOs(stagedOutputs)
}
