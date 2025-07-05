package core

import (
	"bytes"
	"crypto/sha256"
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

func (p *POW) GetData(nonce int32) []byte {
	// Converts POW to bytes
	return bytes.Join([][]byte{
		p.Block.GetData(),
		GetHexValue(int64(nonce)),
		GetHexValue(int64(p.Difficulty)),
	}, []byte{})
}

func (p *POW) Run() (int32, []byte) {
	// Return the nonce and hash
	var intHash big.Int
	var hash [32]byte

	nonce := 0
	for ; nonce < math.MaxInt64; nonce++ {
		data := p.GetData(int32(nonce))
		hash = sha256.Sum256(data)
		intHash.SetBytes(hash[:])

		if intHash.Cmp(p.Target) == -1 { // Check if we found a value that is smaller than target
			break
		}
	}
	return int32(nonce), hash[:]
}

func (p *POW) Validate() bool {
	//  Simply reversing the POW.Run() function
	data := p.GetData(p.Block.Nonce)
	hash := sha256.Sum256(data)
	var intHash big.Int
	intHash.SetBytes(hash[:])
	return intHash.Cmp(p.Target) == -1
}
