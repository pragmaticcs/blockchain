package storage

import (
	"github.com/pragmaticcs/blockchain/blockchain"
	"github.com/dgraph-io/badger"
)

var (
	blockPrefix = []byte("block-")
	utxoPrefix  = []byte("utxo-")
)

type BadgerStore struct {
	// Implements Store interface
	DB *badger.DB
}

func NewBadgerStore(path string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &BadgerStore{
		DB: db,
	}, nil
}

func (b *BadgerStore) Close() error {
	return b.DB.Close()
}

func (b *BadgerStore) GetAll() ([]Command, error) {
	var blocks []*blockchain.Block
	err := b.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(val []byte) error {
				blocks = append(blocks)
			})
		}
	})
}
