package blockchain

import (
	"github.com/dgraph-io/badger/v4"
)

type BadgerPrefix struct {
	Prefix string
}

func (b BadgerPrefix) Len() int {
	return len(b.Prefix)
}

func (b BadgerPrefix) ConstructKey(key []byte) []byte {
	return append([]byte(b.Prefix), key...)
}

func (b BadgerPrefix) DeconstructKey(key []byte) []byte {
	return key[(b.Len() - 1):]
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

func (b *BadgerStore) ForEachBlock(fn func(block *Block) error) error {
	return b.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(blockPrefix.Prefix)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			var block *Block
			err := item.Value(func(val []byte) error {
				value, err := b.BlockEncoder.Decode(val)
				if err != nil {
					return err
				}
				block = value
				return nil
			})
			if err != nil {
				return err
			}

			if err := fn(block); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *BadgerStore) DeleteAllUTXOs() error {
	return b.DeletePrefix(utxoPrefix)
}

func (b *BadgerStore) DeletePrefix(Prefix BadgerPrefix) error {
	return b.DB.DropPrefix([]byte(Prefix.Prefix))
}

// func (b *BadgerStore) PutUTXO(txID TransactionID, utxo *TxOutputs) error {
// 	value, err := b.UtxoEncoder.Encode(utxo)
// 	if err != nil {
// 		return err
// 	}
// 	key := utxoPrefix.ConstructKey(txID[:])
// 	return b.DB.Update(func(txn *badger.Txn) error {
// 		return txn.Set(key, value)
// 	})
// }

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

func (b *BadgerStore) ForEachUTXO(fn func(txID TransactionID, utxo *TxOutputs) error) error {
	return b.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(utxoPrefix.Prefix)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			key := make([]byte, item.KeySize())
			item.KeyCopy(key)
			key = utxoPrefix.DeconstructKey(key)
			txID := TransactionID(key)

			var utxos *TxOutputs
			err := item.Value(func(val []byte) error {
				value, err := b.UtxoEncoder.Decode(val)
				if err != nil {
					return err
				}
				utxos = value
				return nil
			})
			if err != nil {
				return err
			}

			if err := fn(txID, utxos); err != nil {
				return err
			}
		}
		return nil
	})
}

// func (b *BadgerStore) GetAllUTXOs() (UTXOs, error) {
// 	allUtxos := make(UTXOs)

// 	err := b.ForEachUTXO(func(txID TransactionID, utxo *TxOutputs) error {
// 		allUtxos[txID] = *utxo
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return allUtxos, nil
// }

func (b *BadgerStore) PutUTXOs(utxos UTXOs) error {
	return b.DB.Update(func(txn *badger.Txn) error {
		for txID, utxo := range utxos {
			key := utxoPrefix.ConstructKey(txID[:])
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
