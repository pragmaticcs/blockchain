package p2p

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	xOpts := TCPTransportOpts{
		ListenAddr: "localhost:3000",
		SeedPeers:  nil,
		Handshake:  NopHandshakeFunc,
	}
	yOpts := TCPTransportOpts{
		ListenAddr: "localhost:3333",
		SeedPeers:  nil,
		Handshake:  NopHandshakeFunc,
	}
	x := NewTCPTransport(xOpts)
	y := NewTCPTransport(yOpts)
	defer x.Close()
	defer y.Close()
	go func() {
		err := x.ListenAndServe()
		assert.NoError(t, err)
	}()
	go func() {
		err := y.ListenAndServe()
		assert.NoError(t, err)
	}()
	time.Sleep(time.Second * 3)
}
