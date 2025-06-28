package storage

import "github.com/pragmaticcs/blockchain/blockchain"

type Command struct {
	key   []byte
	value []byte
}

type Store interface {
	PutBlock(*blockchain.Block) error
	GetBlock([]byte) *blockchain.Block
	GetCurrentBlock() ([]byte, error)
}
