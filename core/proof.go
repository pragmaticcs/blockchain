package core

import (
	"context"
	"math"
	"math/big"
)

type POW struct {
	// Proof of Work required to mine the block
	Target     *big.Int
	Difficulty uint32
}

func NewPOW(difficulty uint32) *POW {
	// Creating a new Proof of Work with given block and difficulty
	target := big.NewInt(1)
	return &POW{
		Target:     target.Lsh(target, 256-uint(difficulty)),
		Difficulty: difficulty,
	}
}

func (p *POW) Run(ctx context.Context, b *Block) bool {
	// Return the hash of the block
	var intHash big.Int

	b.Difficulty = p.Difficulty

	for b.Nonce = 0; b.Nonce < math.MaxInt32; b.Nonce++ {
		select {
		case <-ctx.Done():
			return false
		default:
			hash := b.GetHash()
			intHash.SetBytes(hash[:])

			if intHash.Cmp(p.Target) == -1 { // Check if we found a value that is smaller than target
				return true
			}
		}

	}
	return false
}

func (p *POW) Validate(b *Block) bool {
	//  Simply reversing the POW.Run() function
	hash := b.GetHash()
	var intHash big.Int
	intHash.SetBytes(hash[:])
	return intHash.Cmp(p.Target) == -1
}
