package p2p

type RpcCommand uint16

const (
	// For peer discovery and connections
	RpcCommandVersion RpcCommand = iota
	RpcCommandAddr               // Msg containing addresses of peers
	RpcCommandGetAddr            // Msg requesting addressess of peers

	// For syncing
	RpcCommandGetBlocks // Msg requesting hashes of blocks
	RpcCommandInv       // Msg announcing items it has using hashes
	RpcCommandGetData   // Msg requesting full contents of items
	RpcCommandBlock     // Msg containing full contents of a block

	// For real-time data sharing (Gossip protocol)
	RpcCommandTx // Message containing a single transaction
)

type RPC struct {
	// RPC is the wrapper for parsed messages
	From    string
	Command RpcCommand
	Payload []byte
}
