package storage

import (
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
