package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase58(t *testing.T) {
	s := EncodeToBase58([]byte("hello"))
	res := DecodeFromBase58(s)
	assert.Equal(t, "hello", string(res))
}
