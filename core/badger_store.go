package core

import (
	"github.com/pragmaticcs/blockchain/utils"
	"github.com/dgraph-io/badger/v4"
)

type BadgerPrefix struct {
	Prefix []byte
}

func NewBadgerPrefix(prefix []byte) BadgerPrefix {
	return BadgerPrefix{Prefix: prefix}
}

func (b BadgerPrefix) Len() int {
	return len(b.Prefix)
}

func (b BadgerPrefix) ConstructKey(key []byte) []byte {
	return append(b.Prefix, key...)
}

func (b BadgerPrefix) DeconstructKey(key []byte) []byte {
	return key[b.Len():]
}

var (
	blockPrefix       = NewBadgerPrefix([]byte("block-"))
	utxoPrefix        = NewBadgerPrefix([]byte("utxo-"))
	blockchainHeadKey = []byte("blockchainhead-")
)

type BadgerStore struct {
	// Implements Store interface
	DB *badger.DB
}

func NewBadgerStore(path string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(path).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &BadgerStore{
		DB: db,
	}, nil
}

func (b *BadgerStore) Exists(hash BlockHash) bool {
	exists := true
	b.DB.View(func(txn *badger.Txn) error {
		key := blockPrefix.ConstructKey(hash)
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			exists = false
		}
		return nil
	})
	return exists
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
	return utils.DecodeInGob[Block](value)
}

func (b *BadgerStore) PutBlock(block *Block) error {
	value, err := utils.EncodeInGob(block)
	if err != nil {
		return err
	}
	key := blockPrefix.ConstructKey(block.GetHash())
	return b.DB.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (b *BadgerStore) GetBlockchainHead() (*Block, error) {
	var value []byte
	err := b.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(blockchainHeadKey)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(value)
		return err
	})
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return utils.DecodeInGob[Block](value)
}

func (b *BadgerStore) SetBlockchainHead(block *Block) error {
	value, err := utils.EncodeInGob(block)
	if err != nil {
		return err
	}
	return b.DB.Update(func(txn *badger.Txn) error {
		return txn.Set(blockchainHeadKey, value)
	})
}

func (b *BadgerStore) DeleteAllUTXOs() error {
	return b.DeletePrefix(utxoPrefix)
}

func (b *BadgerStore) DeletePrefix(Prefix BadgerPrefix) error {
	return b.DB.DropPrefix([]byte(Prefix.Prefix))
}

func (b *BadgerStore) PutUTXOs(utxos UTXOs) error {
	return b.DB.Update(func(txn *badger.Txn) error {
		for outPoint, utxo := range utxos {
			key, err := EncodeOutpoint(&outPoint)
			if err != nil {
				return err
			}
			key = utxoPrefix.ConstructKey(key)
			value, err := utils.EncodeInGob(&utxo)
			if err != nil {
				return err
			}
			err = txn.Set(key, value)
			if err != nil {
				return err
			}
		}
		return nil
	})
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

func (b *BadgerStore) GetUTXO(outpoint OutPoint) (*TxOutput, error) {
	var value []byte
	err := b.DB.View(func(txn *badger.Txn) error {
		outPointData, err := EncodeOutpoint(&outpoint)
		if err != nil {
			return err
		}
		key := utxoPrefix.ConstructKey(outPointData)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(value)
		return err
	})
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return utils.DecodeInGob[TxOutput](value)
}

func (b *BadgerStore) PutDeleteUTXOs(put UTXOs, delete []OutPoint) error {
	// This method combines adding and deleting values
	// in one atomic transaction
	return b.DB.Update(func(txn *badger.Txn) error {
		for outPoint, utxo := range put {
			key, err := EncodeOutpoint(&outPoint)
			if err != nil {
				return err
			}
			key = utxoPrefix.ConstructKey(key)
			value, err := utils.EncodeInGob(&utxo)
			if err != nil {
				return err
			}
			err = txn.Set(key, value)
			if err != nil {
				return err
			}
		}
		for _, outPoint := range delete {
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
			if err != nil {
				return err
			}

			var utxo *TxOutput
			err = item.Value(func(val []byte) error {
				value, err := utils.DecodeInGob[TxOutput](val)
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

func (b *BadgerStore) Close() error {
	return b.DB.Close()
}
