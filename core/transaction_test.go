package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoinbase(t *testing.T) {
	coinbase := NewCoinbase(1000000, []byte("my public key"))
	assert.True(t, coinbase.IsCoinbase())
}
