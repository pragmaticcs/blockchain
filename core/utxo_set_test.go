package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutpointEncoding(t *testing.T) {
	txID := []byte("12345678123456781234567812345678")

	outpoint := NewOutPoint(TransactionID(txID), 5)
	enc, err := EncodeOutpoint(&outpoint)
	assert.NoError(t, err)
	dec, err := DecodeOutpoint(enc)
	assert.NoError(t, err)
	assert.EqualValues(t, outpoint, *dec)
}
