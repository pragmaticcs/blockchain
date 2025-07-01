package blockchain

import (
	"github.com/pragmaticcs/blockchain/crypto"
	"github.com/pragmaticcs/blockchain/types"
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

// Implements the types.iterator interface
type BlockchainIterator struct {
	CurrHash BlockHash
	Store    Store
}

func (b *BlockchainIterator) Next() (*Block, bool) {
	if b.CurrHash == nil {
		return nil, false
	}
	currBlock, err := b.Store.GetBlock(b.CurrHash)
	if err != nil {
		return nil, false
	}
	b.CurrHash = currBlock.PrevHash
	return currBlock, true
}

func (b *BlockchainIterator) Close() {
	// Currently doesn't do anything
}

func (b *Blockchain) ToIterator() types.Iterator[*Block] {
	return &BlockchainIterator{
		CurrHash: b.Head,
		Store:    b.Store,
	}
}

func (b *Blockchain) AddBlock(transactions []*Transaction) (*Block, error) {
	block := NewBlock(b.Head, transactions)
	// Run the proof of work
	pow := NewPOW(block, b.CurrDifficulty)

	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash
	b.Head = hash

	// Add the mined block to the store
	err := b.Store.PutBlock(block)
	if err != nil {
		return nil, err
	}
	return block, nil
}
