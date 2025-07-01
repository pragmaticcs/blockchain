package blockchain

import (
	"github.com/pragmaticcs/blockchain/crypto"
	"github.com/pragmaticcs/blockchain/types"
)

type BlockChain struct {
	CurrDifficulty uint
	Blocks         []*Block
}

type GenesisConfig struct {
	Amount int
	To     crypto.PubKey
	Sig    crypto.Signature
}

func NewBlockChain(difficulty int, genesisConfig GenesisConfig) (*BlockChain, error) {
	genesis, err := CreateGenesisBlock(genesisConfig.Amount, genesisConfig.To)
	if err != nil {
		return nil, err
	}
	return &BlockChain{
		CurrDifficulty: uint(difficulty),
		Blocks: []*Block{
			genesis,
		},
	}, nil
}

type BlockChainIterator struct {
	Blocks    []*Block
	CurrIndex int
}

func (b *BlockChain) ToIterator() *BlockChainIterator {
	return &BlockChainIterator{
		Blocks:    b.Blocks,
		CurrIndex: len(b.Blocks) - 1,
	}
}

func (b *BlockChainIterator) Prev() *Block {
	// Goes from the most recent block to the genesis block
	defer func() { b.CurrIndex -= 1 }()
	if b.CurrIndex < 0 {
		return nil
	}
	return b.Blocks[b.CurrIndex]
}

func (b *BlockChain) PrevBlock() *Block {
	return b.Blocks[len(b.Blocks)-1]
}

func (b *BlockChain) AddBlock(transactions []*Transaction) *Block {
	// Add a block to the chain of blocks (Duh)
	block := NewBlock(b.PrevBlock().Hash, transactions)
	pow := NewPOW(block, b.CurrDifficulty)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash

	b.Blocks = append(b.Blocks, block)

	return block
}

// NEW Version that uses a store
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

func (b *Blockchain) ToBlockIterator() types.Iterator[*Block] {
	return &BlockchainIterator{
		CurrHash: b.Head,
		Store:    b.Store,
	}
}

func (b *Blockchain) AddBlock(transactions []*Transaction) *Block {
	block := NewBlock(b.Head, transactions)
	b.Store.PutBlock(block)
	pow := NewPOW(block, b.CurrDifficulty)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash
	b.Head = block.Hash
	return block
}
