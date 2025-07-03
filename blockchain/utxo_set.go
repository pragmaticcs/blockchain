package blockchain

import "sync"

type TxOutputs struct {
	Outputs []TxOutput
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

// func (u *UTXOset) Reindex() error {
// 	u.Lock()
// 	defer u.Unlock()
// 	utxos, err := u.Blockchain.FindUtxoSet()
// 	if err != nil {
// 		return err
// 	}
// 	u.Store.DeleteAllUTXOs()
// 	for txID, out := range utxos {
// 		err := u.Store.PutUTXO(txID, &TxOutputs{
// 			Outputs: out,
// 		})
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func (u *UTXOset) Update(block *Block) error {
	u.Lock()
	defer u.Unlock()
	return nil
}
