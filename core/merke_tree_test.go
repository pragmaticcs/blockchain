package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMerkleTree(t *testing.T) {
	txIDs := []TransactionID{
		TransactionID([]byte("1234567812345678123456781234567812345678123456781234567812345678")),
		TransactionID([]byte("abcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefgh")),
		TransactionID([]byte("1111111122222222333333334444444455555555666666667777777788888888")),
	}

	tree, err := NewMerkleTree(txIDs)
	assert.NoError(t, err)
	proof, err := tree.GetProof(txIDs[0])
	assert.NoError(t, err)
	validate := proof.Validate(txIDs[0], tree.Root)
	assert.Equal(t, true, validate)

	fakeTxID := TransactionID([]byte("1111111111111111111111111111111111111111111111111111111111111111"))
	validate = proof.Validate(fakeTxID, tree.Root)
	assert.Equal(t, false, validate)
}
