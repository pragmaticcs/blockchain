package p2p

import (
	"github.com/pragmaticcs/blockchain/core"
	"github.com/pragmaticcs/blockchain/wallet"
)

type NodeOpts struct {
	Port          string
	SeedNodes     []string
	WalletFileDir string
}

func DefaultNodeOpts() NodeOpts {
	return NodeOpts{
		Port:          "3000",
		SeedNodes:     nil,
		WalletFileDir: "", // Save to current directory
	}
}

type Node struct {
	StateManager *core.StateManager
	Mempool      *Mempool
	Transport    *TCPTransport // TODO: Replace with Transport interface

	// Channels
	quit chan struct{}
}

func NewNode(opts NodeOpts, b *core.Blockchain, w *wallet.Wallets) *Node {
	transportOpts := TCPTransportOpts{
		ListenAddr: opts.Port,
		SeedPeers:  opts.SeedNodes,
	}
	transport := NewTCPTransport(transportOpts)

	stateManager := core.NewStateManager(b, w)
	return &Node{
		StateManager: stateManager,
		Mempool:      nil,
		Transport:    transport,
	}
}

func (n *Node) Start() error {
	defer n.Stop()

	err := n.Transport.Start()
	if err != nil {
		return err
	}

	defer n.Transport.Close()

	n.MainLoop()
	return nil
}

func (n *Node) Stop() {
	// Handles graceful shutdown
	n.StateManager.Close()
}

func (n *Node) sendQuitSignal() {
	// Sends a quit signal
	n.quit <- struct{}{}
}

func (n *Node) MainLoop() {
	for {
		select {
		case rpc := <-n.Transport.RpcCh:
			n.HandleRpc(rpc)
		case _ = <-n.quit:
			return
		}
	}
}

func (n *Node) HandleRpc(rpc RPC) error {
	switch rpc.Command {
	case RpcCommandVersion:
		return n.HandleMessageVersion(rpc)
	case RpcCommandBlock:
	case RpcCommandAddr:
	case RpcCommandTx:
	default:
		return nil
	}
	return nil
}

func (n *Node) HandleMessageVersion(rpc RPC) error {
	encoder := core.GobEncoder[MessageVersion]{}
	message, err := encoder.Decode(rpc.Payload)
	if err != nil {
		return err
	}
	if message.ProtocolVersion != core.CurrentVersion {
		// TODO: Handle this
	}	
	if message.LatestBlockHeight > n.StateManager.GetLatestBlockHeight()
	return nil
}
