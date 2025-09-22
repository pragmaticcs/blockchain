package node

import (
	"sync"

	"github.com/pragmaticcs/blockchain/core"
)

type Mempool struct {
	transactions map[core.TransactionID]*core.Transaction
	lock         sync.RWMutex
	maxSize      int
}

func NewMempool(maxSize int) *Mempool {
	return &Mempool{
		transactions: make(map[core.TransactionID]*core.Transaction),
		lock:         sync.RWMutex{},
		maxSize:      maxSize,
	}
}

func (m *Mempool) GetTransactions() []*core.Transaction {
	m.lock.Lock()
	defer m.lock.Unlock()
	txs := make([]*core.Transaction, 0, len(m.transactions))

	for _, tx := range m.transactions {
		txs = append(txs, tx)
	}
	return txs
}

func (m *Mempool) GetTransaction(txID core.TransactionID) (*core.Transaction, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	tx, ok := m.transactions[core.TransactionID(txID)]
	return tx, ok
}

func (m *Mempool) Clear(newBlock *core.Block) {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, tx := range newBlock.Transactions {
		delete(m.transactions, tx.ID)
	}
}

func (m *Mempool) TxExists(txID []byte) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	_, ok := m.transactions[core.TransactionID(txID)]
	return ok
}

func (m *Mempool) AddTx(tx *core.Transaction) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if len(m.transactions)+1 > m.maxSize {
		return false
	}
	m.transactions[tx.ID] = tx
	return true
}

func (m *Mempool) RemoveTx(txID core.TransactionID) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.transactions, txID)
}
