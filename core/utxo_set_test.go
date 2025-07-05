package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutpointEncoding(t *testing.T) {
	handle := func(err error) {
		if err != nil {
			t.Error(err)
		}
	}
	txID := []byte("12345678123456781234567812345678")

	outpoint := NewOutPoint(TransactionID(txID), 5)
	enc, err := EncodeOutpoint(&outpoint)
	handle(err)
	dec, err := DecodeOutpoint(enc)
	handle(err)
	assert.EqualValues(t, outpoint, *dec)
}
