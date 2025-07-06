package core

import (
	"math"
	"math/big"
)

type POW struct {
	// Proof of Work required to mine the block
	Block      *Block
	Target     *big.Int
	Difficulty uint
}

func NewPOW(b *Block, difficulty uint) *POW {
	// Creating a new Proof of Work with given block and difficulty
	target := big.NewInt(1)
	return &POW{
		Block:      b,
		Target:     target.Lsh(target, 256-difficulty),
		Difficulty: difficulty,
	}
}

func (p *POW) Run() bool {
	// Return the hash of the block
	var intHash big.Int

	mined := false
	for p.Block.Nonce = 0; p.Block.Nonce < math.MaxInt32; p.Block.Nonce++ {
		hash := p.Block.GetHash()
		intHash.SetBytes(hash[:])

		if intHash.Cmp(p.Target) == -1 { // Check if we found a value that is smaller than target
			mined = true
			break
		}
	}
	return mined
}

func (p *POW) Validate() bool {
	//  Simply reversing the POW.Run() function
	hash := p.Block.GetHash()
	var intHash big.Int
	intHash.SetBytes(hash[:])
	return intHash.Cmp(p.Target) == -1
}
