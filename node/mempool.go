package node

import (
	"sync"

	"github.com/pragmaticcs/blockchain/core"
)

type Mempool struct {
	transactions map[core.TransactionID]*core.Transaction
	lock         sync.RWMutex
}

func NewMempool() *Mempool {
	return &Mempool{
		transactions: make(map[core.TransactionID]*core.Transaction),
		lock:         sync.RWMutex{},
	}
}

func (m *Mempool) AddTx(tx *core.Transaction) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.transactions[tx.ID] = tx
}

func (m *Mempool) RemoveTx(txID core.TransactionID) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.transactions, txID)
}
