package p2p

type RPC struct {
	From    string
	Payload []byte
}

// Handles the transport (TCP, UDP)
type Transport interface {
}
