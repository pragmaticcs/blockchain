package core

import (
	"github.com/pragmaticcs/blockchain/crypto"
)

type GenesisConfig struct {
	Amount int
	To     crypto.PubKey
}

type Blockchain struct {
	Store          Store
	Head           BlockHash
	CurrDifficulty uint32
}

func NewBlockchain(difficulty int, genesisConfig GenesisConfig, store Store) (*Blockchain, error) {
	genesis, err := CreateGenesisBlock(genesisConfig.Amount, genesisConfig.To)
	if err != nil {
		return nil, err
	}
	store.PutBlock(genesis)
	return &Blockchain{
		Store:          store,
		Head:           genesis.GetHash(),
		CurrDifficulty: uint32(difficulty),
	}, nil
}

func (b *Blockchain) Iterate() func(func(*Block) bool) {
	return func(yield func(*Block) bool) {
		currHash := b.Head
		for currHash != nil {
			currBlock, err := b.Store.GetBlock(currHash)
			if err != nil {
				return
			}
			if !yield(currBlock) {
				return
			}
			currHash = currBlock.PrevHash
		}
	}
}

func (b *Blockchain) NewBlock(transactions []*Transaction) (*Block, error) {
	return NewBlock(b.Head, transactions)
}

func (b *Blockchain) NewPOW() *POW {
	return NewPOW(b.CurrDifficulty)
}

func (b *Blockchain) AddBlock(block *Block) error {
	b.Head = block.GetHash()
	return b.Store.PutBlock(block)
}
