package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

type BadgerPrefix struct {
	Prefix []byte
}

func (b BadgerPrefix) Len() int {
	return len(b.Prefix)
}

func (b BadgerPrefix) ConstructKey(key []byte) []byte {
	return append(b.Prefix, key...)
}

func (b BadgerPrefix) DeconstructKey(key []byte) []byte {
	return key[(b.Len() - 1):]
}

var (
	blockPrefix = BadgerPrefix{Prefix: []byte("block-")}
	utxoPrefix  = BadgerPrefix{Prefix: []byte("utxo-")}
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
		UtxoEncoder:  GobEncoder[TxOutput]{},
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

func (b *BadgerStore) DeleteAllUTXOs() error {
	return b.DeletePrefix(utxoPrefix)
}

func (b *BadgerStore) DeletePrefix(Prefix BadgerPrefix) error {
	return b.DB.DropPrefix([]byte(Prefix.Prefix))
}

func (b *BadgerStore) DeleteUTXOs(outPoints []OutPoint) error {
	return b.DB.Update(func(txn *badger.Txn) error {
		for _, outPoint := range outPoints {
			outPointData, err := EncodeOutpoint(&outPoint)
			if err != nil {
				return err
			}
			key := utxoPrefix.ConstructKey(outPointData)
			err = txn.Delete(key)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *BadgerStore) ForEachUTXO(fn func(outPoint OutPoint, out TxOutput) error) error {
	return b.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = utxoPrefix.Prefix

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()

			outPointData := utxoPrefix.DeconstructKey(key)

			outPoint, err := DecodeOutpoint(outPointData)
			fmt.Println(outPoint.TxID)
			if err != nil {
				return err
			}

			var utxo *TxOutput
			err = item.Value(func(val []byte) error {
				value, err := b.UtxoEncoder.Decode(val)
				if err != nil {
					return err
				}
				utxo = value
				return nil
			})
			if err != nil {
				return err
			}

			if err := fn(*outPoint, *utxo); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *BadgerStore) PutUTXOs(utxos UTXOs) error {
	return b.DB.Update(func(txn *badger.Txn) error {
		for outPoint, utxo := range utxos {
			key, err := EncodeOutpoint(&outPoint)
			if err != nil {
				return err
			}
			key = utxoPrefix.ConstructKey(key)
			value, err := b.UtxoEncoder.Encode(&utxo)
			if err != nil {
				return err
			}
			txn.Set(key, value)
		}
		return nil
	})
}

func (b *BadgerStore) Close() error {
	return b.DB.Close()
}
