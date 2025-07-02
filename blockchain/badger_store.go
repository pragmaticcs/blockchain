package blockchain

import (
	"github.com/pragmaticcs/blockchain/types"
	"github.com/dgraph-io/badger/v4"
)

type BadgerPrefix struct {
	Prefix string
}

func (b BadgerPrefix) ConstructKey(key []byte) []byte {
	return append([]byte(b.Prefix), key...)
}

var (
	blockPrefix = BadgerPrefix{Prefix: "block-"}
	utxoPrefix  = BadgerPrefix{Prefix: "utxo-"}
)

type BadgerStore struct {
	// Implements Store interface
	DB           *badger.DB
	BlockEncoder BlockEncoder
	UtxoEncoder  UtxoEncoder
}

func NewBadgerStore(path string) (*BadgerStore, error) {
	opts := badger.DefaultOptions("").WithInMemory(true).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &BadgerStore{
		DB:           db,
		BlockEncoder: GobEncoder[Block]{},
		UtxoEncoder:  GobEncoder[TxOutputs]{},
	}, nil
}

func (b *BadgerStore) GetBlock(hash BlockHash) (*Block, error) {
	var value []byte
	err := b.DB.View(func(txn *badger.Txn) error {
		key := blockPrefix.ConstructKey(hash)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(value)
		return err
	})
	if err != nil {
		return nil, err
	}
	return b.BlockEncoder.Decode(value)
}

func (b *BadgerStore) PutBlock(block *Block) error {
	value, err := b.BlockEncoder.Encode(block)
	if err != nil {
		return err
	}
	key := blockPrefix.ConstructKey(block.Hash)
	return b.DB.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (b *BadgerStore) DeletePrefix(Prefix BadgerPrefix) error {
	return b.DB.DropPrefix([]byte(Prefix.Prefix))
}

func (b *BadgerStore) PutUTXO(txID TransactionID, utxo *TxOutputs) error {
	value, err := b.UtxoEncoder.Encode(utxo)
	if err != nil {
		return err
	}
	key := utxoPrefix.ConstructKey(txID[:])
	return b.DB.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (b *BadgerStore) GetUTXO(txID TransactionID) (*TxOutputs, error) {
	var value []byte
	err := b.DB.View(func(txn *badger.Txn) error {
		key := utxoPrefix.ConstructKey(txID[:])
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(value)
		return err
	})
	if err != nil {
		return nil, err
	}
	return b.UtxoEncoder.Decode(value)
}

func (b *BadgerStore) DeleteUTXO(txID TransactionID) error {
	key := utxoPrefix.ConstructKey(txID[:])
	return b.DB.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
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
	opts.PrefetchValues = false
	opts.Prefix = []byte(blockPrefix.Prefix)

	it := txn.NewIterator(opts)
	it.Rewind()

	return &BadgerBlockIterator{txn: txn, it: it}
}

type BadgerUtxosIterator struct {
	txn     *badger.Txn
	it      *badger.Iterator
	encoder UtxoEncoder
}

func (it *BadgerUtxosIterator) Next() (*TxOutputs, bool) {
	if !it.it.Valid() {
		return nil, false
	}

	var utxos *TxOutputs
	item := it.it.Item()
	err := item.Value(func(val []byte) error {
		value, err := it.encoder.Decode(val)
		if err != nil {
			return err
		}
		utxos = value
		return nil
	})
	if err != nil {
		// TODO: handle this error better
		return nil, false
	}
	it.it.Next()
	return utxos, true
}

func (it *BadgerUtxosIterator) Close() {
	it.it.Close()
}

func (s *BadgerStore) ToUtxoIterator() types.Iterator[*TxOutputs] {
	txn := s.DB.NewTransaction(false)
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	opts.Prefix = []byte(utxoPrefix.Prefix)

	it := txn.NewIterator(opts)
	it.Rewind()

	return &BadgerUtxosIterator{txn: txn, it: it}
}
