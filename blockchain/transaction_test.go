package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoinbase(t *testing.T) {
	coinbase, err := NewCoinbase(1000000, []byte("my public key"))
	if err != nil {
		t.Error(err)
	}
	assert.True(t, coinbase.IsCoinbase())
}
