package p2p

import "io"

// Handles the transport (TCP, UDP, gRPC)
type Transport interface {
}

type Peer interface {
	io.ReadWriter
	Send([]byte) error
	GetAddr() string
	Close() error
}

type HandshakeFunc func(Peer) error
