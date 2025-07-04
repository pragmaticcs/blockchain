package blockchain

import (
	"github.com/pragmaticcs/blockchain/crypto"
)

type GenesisConfig struct {
	Amount int
	To     crypto.PubKey
	Sig    crypto.Signature
}

type Blockchain struct {
	Store          Store
	Head           BlockHash
	CurrDifficulty uint
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
		CurrDifficulty: uint(difficulty),
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

func (b *Blockchain) AddBlock(transactions []*Transaction) (*Block, error) {
	block := NewBlock(b.Head, transactions)
	// Run the proof of work
	pow := NewPOW(block, b.CurrDifficulty)

	block.Nonce, block.Hash = pow.Run()

	b.Head = block.Hash

	// Add the mined block to the store
	err := b.Store.PutBlock(block)
	if err != nil {
		return nil, err
	}
	return block, nil
}
