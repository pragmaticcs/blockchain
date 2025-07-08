package p2p

type RPC struct {
	From    string
	Command RpcCommand
	Payload []byte
}
