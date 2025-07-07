package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoinbase(t *testing.T) {
	coinbase, err := NewCoinbase(1000000, []byte("my public key"))
	assert.NoError(t, err)
	assert.True(t, coinbase.IsCoinbase())
}
