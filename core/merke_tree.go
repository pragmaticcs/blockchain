package core

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

type MerkleHash = []byte

type MerkleProofBranch struct {
	Hash   MerkleHash
	IsLeft bool
}

type MerkleProof struct {
	Branches []MerkleProofBranch
}

type MerkleTreeLevel = []MerkleHash

type MerkleTree struct {
	Root   MerkleHash
	Levels []MerkleTreeLevel
}

func NewMerkleTree(txIDs []TransactionID) (*MerkleTree, error) {
	if len(txIDs) == 0 {
		return nil, NewMerkleTreeError("can't create a merkle tree for 0 transactions")
	}
	merkleTree := &MerkleTree{}

	currLevel := make([]MerkleHash, len(txIDs))
	for i, txID := range txIDs {
		currLevel[i] = DoubleSha256(txID[:])
	}
	merkleTree.Levels = append(merkleTree.Levels, currLevel)

	for len(currLevel) > 1 {
		var nextLevel MerkleTreeLevel
		for currLevelIndex := 0; currLevelIndex < len(currLevel); currLevelIndex += 2 {
			left := currLevel[currLevelIndex]
			var right MerkleHash
			if currLevelIndex+1 < len(currLevel) {
				right = currLevel[currLevelIndex+1]
			} else {
				right = left
			}
			nextLevel = append(nextLevel, ConnectMerkleBranches(left, right))
		}
		currLevel = nextLevel
		merkleTree.Levels = append(merkleTree.Levels, currLevel)
	}
	merkleTree.Root = merkleTree.Levels[len(merkleTree.Levels)-1][0]
	return merkleTree, nil
}

func (t *MerkleTree) GetProof(txID TransactionID) (*MerkleProof, error) {
	var branches []MerkleProofBranch

	var currIndex int = -1
	for id, hash := range t.Levels[0] {
		if bytes.Equal(hash, DoubleSha256(txID[:])) {
			currIndex = id
			break
		}
	}
	if currIndex == -1 {
		return nil, NewMerkleTreeError("merkle tree does not contain given transaction id")
	}

	for _, level := range t.Levels[:(len(t.Levels) - 1)] {
		var pair MerkleHash
		pairIsLeft := currIndex%2 != 0
		if pairIsLeft {
			pair = level[currIndex-1]
		} else if currIndex == len(level)-1 {
			pair = level[currIndex]
		} else {
			pair = level[currIndex+1]
		}
		branch := MerkleProofBranch{
			Hash:   pair,
			IsLeft: pairIsLeft,
		}
		branches = append(branches, branch)
		currIndex = currIndex / 2
	}

	return &MerkleProof{
		Branches: branches,
	}, nil
}

func (t *MerkleProof) Validate(txID TransactionID, root MerkleHash) bool {
	currHash := DoubleSha256(txID[:])
	for _, branch := range t.Branches {
		var concat MerkleHash
		if branch.IsLeft {
			concat = append(branch.Hash, currHash...)
		} else {
			concat = append(currHash, branch.Hash...)
		}
		currHash = DoubleSha256(concat)
	}
	return bytes.Equal(currHash, root)
}

func ConnectMerkleBranches(a, b []byte) MerkleHash {
	concat := append(a, b...)
	return DoubleSha256(concat)
}

func DoubleSha256(data []byte) MerkleHash {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}

type MerkleTreeError struct {
	Err string
}

func NewMerkleTreeError(e string) MerkleTreeError {
	return MerkleTreeError{Err: e}
}

func (e MerkleTreeError) Error() string {
	return fmt.Sprintf("Merkle tree error: %s", e.Err)
}
