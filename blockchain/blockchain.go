package blockchain

type BlockChain struct {
	currDifficulty uint
	blocks         []*Block
}

func NewBlockChain(difficulty int) *BlockChain {
	return &BlockChain{
		currDifficulty: uint(difficulty),
		blocks: []*Block{
			CreateGenesisBlock(),
		},
	}
}

func (b *BlockChain) PrevBlock() *Block {
	return b.blocks[len(b.blocks)-1]
}

func (b *BlockChain) AddBlock(data []byte) *Block {
	block := NewBlock(b.PrevBlock().Hash, data)
	pow := NewPOW(block, b.currDifficulty)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	copy(block.Hash[:], hash)

	b.blocks = append(b.blocks, block)
	return block
}
