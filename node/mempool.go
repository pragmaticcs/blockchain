package node

import (
	"sync"

	"github.com/pragmaticcs/blockchain/core"
	"github.com/pragmaticcs/blockchain/crypto"
)

type Mempool struct {
	transactions map[core.TransactionID]*core.Transaction
	lock         sync.RWMutex
	maxSize      int
	// Yeah, bitcoin terms are like this, nothing I can do
	orphans map[core.TransactionID]*core.Transaction
}

func NewMempool(maxSize int) *Mempool {
	return &Mempool{
		transactions: make(map[core.TransactionID]*core.Transaction),
		lock:         sync.RWMutex{},
		maxSize:      maxSize,
		orphans:      make(map[core.TransactionID]*core.Transaction),
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

func (m *Mempool) GetTransactionTable() map[core.TransactionID]*core.Transaction {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.transactions
}

func (m *Mempool) GetReservedOutputs(pubKey crypto.PubKey) (map[core.OutPoint]struct{}, map[core.OutPoint]core.TxOutput) {
	m.lock.Lock()
	defer m.lock.Unlock()
	reserved := make(map[core.OutPoint]struct{})
	available := make(map[core.OutPoint]core.TxOutput)
	for _, tx := range m.transactions {
		if tx.OwnedBy(pubKey) {
			for _, in := range tx.Inputs {
				outPoint := core.NewOutPoint(in.TxID, in.OutID)
				reserved[outPoint] = struct{}{}
			}
			for outId, out := range tx.Outputs {
				outPoint := core.NewOutPoint(tx.ID, int64(outId))
				available[outPoint] = out
			}
		}
	}
	return reserved, available
}

func (m *Mempool) IsChildTx(tx core.Transaction) (bool, []*core.Transaction) {
	m.lock.Lock()
	defer m.lock.Unlock()
	ancestors := []*core.Transaction{}
	for _, in := range tx.Inputs {
		ancestor, ok := m.transactions[in.TxID]
		if ok {
			ancestors = append(ancestors, ancestor)
		}
	}
	return len(ancestors) > 0, ancestors
}

func (m *Mempool) GetTransaction(txID core.TransactionID) (*core.Transaction, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	tx, ok := m.transactions[core.TransactionID(txID)]
	if !ok {
		txOrphan, ok := m.orphans[core.TransactionID(txID)]
		if !ok {
			return nil, false
		}
		return txOrphan, true
	}
	return tx, true
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

func (m *Mempool) txExistUnlocked(txID []byte) bool {
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

func (m *Mempool) AddOrphanTx(tx *core.Transaction) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.orphans[tx.ID] = tx
}

func (m *Mempool) RemoveTx(txID core.TransactionID) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.transactions, txID)
}
