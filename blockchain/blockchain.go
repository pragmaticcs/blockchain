package blockchain

import (
	"sync"

	"github.com/pragmaticcs/blockchain/crypto"
	"github.com/pragmaticcs/blockchain/wallet"
)

type BlockChain struct {
	CurrDifficulty uint
	Blocks         []*Block
	Wallets        *wallet.Wallets

	// Mutexes
	selfLock    sync.RWMutex // Use when modifying the state of blockchain itself
	accountLock sync.RWMutex // Use when modifying the state of accounts/wallets
}

func (b *BlockChain) Lock() {
	b.selfLock.Lock()
}

func (b *BlockChain) Unlock() {
	b.selfLock.Unlock()
}

func (b *BlockChain) LockAccounts() {
	b.accountLock.Lock()
}

func (b *BlockChain) UnlockAccounts() {
	b.accountLock.Unlock()
}

type GenesisConfig struct {
	Amount int
	To     crypto.PubKey
	Sig    crypto.Signature
}

func NewBlockChain(difficulty int, genesisConfig GenesisConfig, wallets *wallet.Wallets) (*BlockChain, error) {
	genesis, err := CreateGenesisBlock(genesisConfig.Amount, genesisConfig.To)
	if err != nil {
		return nil, err
	}
	return &BlockChain{
		CurrDifficulty: uint(difficulty),
		Blocks: []*Block{
			genesis,
		},
		Wallets: wallets,
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
	copy(block.Hash[:], hash)

	b.Blocks = append(b.Blocks, block)

	return block
}

func (b *BlockChain) GetBalance(addr wallet.Address) (int, error) {
	wallet, err := b.Wallets.GetWallet(addr)
	if err != nil {
		return -1, err
	}
	pubKeyHash, err := crypto.HashPublicKey(wallet.PublicKey)
	if err != nil {
		return -1, err
	}
	return b.FindSpendableAmount(pubKeyHash)

}
