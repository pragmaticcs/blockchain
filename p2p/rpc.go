package p2p

type RpcCommand uint16

const (
	RpcCommandVersion RpcCommand = iota
	RpcCommandAddr
	RpcCommandInv
	RpcCommandBlock
	RpcCommandTx
)

type RPC struct {
	// RPC is the wrapper for parsed messages
	From    string
	Command RpcCommand
	Payload []byte
}
