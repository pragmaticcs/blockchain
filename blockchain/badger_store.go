package blockchain

import (
	"github.com/pragmaticcs/blockchain/types"
	"github.com/dgraph-io/badger/v4"
)

var (
	blockPrefix = []byte("block-")
	utxoPrefix  = []byte("utxo-")
)

type BadgerStore struct {
	// Implements Store interface
	DB      *badger.DB
	Encoder BlockEncoder
}

func NewBadgerStore(path string, encoder BlockEncoder) (*BadgerStore, error) {
	opts := badger.DefaultOptions("").WithInMemory(true).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &BadgerStore{
		DB:      db,
		Encoder: encoder,
	}, nil
}

func (b *BadgerStore) GetBlock(hash BlockHash) (*Block, error) {
	var value []byte
	err := b.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(hash)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(value)
		return err
	})
	if err != nil {
		return nil, err
	}
	return b.Encoder.Decode(value)
}

func (b *BadgerStore) PutBlock(block *Block) error {
	value, err := b.Encoder.Encode(block)
	if err != nil {
		return err
	}
	return b.DB.Update(func(txn *badger.Txn) error {
		return txn.Set(block.Hash, value)
	})
}

func (b *BadgerStore) Close() error {
	return b.DB.Close()
}

type BadgerBlockIterator struct {
	txn     *badger.Txn
	it      *badger.Iterator
	encoder BlockEncoder
}

func (it *BadgerBlockIterator) Next() (*Block, bool) {
	if !it.it.Valid() {
		return nil, false
	}

	var block *Block
	item := it.it.Item()
	err := item.Value(func(val []byte) error {
		b, err := it.encoder.Decode(val)
		if err != nil {
			return err
		}
		block = b
		return nil
	})
	if err != nil {
		// TODO: handle this error better
		return nil, false
	}
	it.it.Next()
	return block, true
}

func (it *BadgerBlockIterator) Close() {
	it.it.Close()
}

func (s *BadgerStore) ToBlockIterator() types.Iterator[*Block] {
	txn := s.DB.NewTransaction(false)
	opts := badger.DefaultIteratorOptions
	opts.Prefix = blockPrefix

	it := txn.NewIterator(opts)
	it.Rewind()

	return &BadgerBlockIterator{txn: txn, it: it}
}
