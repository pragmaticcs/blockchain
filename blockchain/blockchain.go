package blockchain

type BlockChain struct {
	currDifficulty uint
	blocks         []*Block
}

func NewBlockChain(difficulty int) (*BlockChain, error) {
	genesis, err := CreateGenesisBlock()
	if err != nil {
		return nil, err
	}
	return &BlockChain{
		currDifficulty: uint(difficulty),
		blocks: []*Block{
			genesis,
		},
	}, nil
}

func (b *BlockChain) PrevBlock() *Block {
	return b.blocks[len(b.blocks)-1]
}

func (b *BlockChain) AddBlock(transactions []*Transaction) *Block {
	block := NewBlock(b.PrevBlock().Hash, transactions)
	pow := NewPOW(block, b.currDifficulty)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	copy(block.Hash[:], hash)

	b.blocks = append(b.blocks, block)
	return block
}
