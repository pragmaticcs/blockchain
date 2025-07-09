package core

import (
	"fmt"
	"log"

	"github.com/pragmaticcs/blockchain/p2p"
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
	StateManager *StateManager
	Mempool      *Mempool
	Transport    *p2p.TCPTransport // TODO: Replace with Transport interface

	// Channels
	quit chan struct{}
}

func NewNode(opts NodeOpts, b *Blockchain, w *wallet.Wallets) *Node {
	var node Node

	transportOpts := p2p.TCPTransportOpts{
		ListenAddr: opts.Port,
		SeedPeers:  opts.SeedNodes,
		Handshake:  node.Handshake,
	}
	transport := p2p.NewTCPTransport(transportOpts)

	stateManager := NewStateManager(b, w)

	node.Transport = transport
	node.StateManager = stateManager
	node.quit = make(chan struct{})
	node.Mempool = nil
	return &node
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

func (n *Node) HandleRpc(rpc p2p.RPC) error {
	switch rpc.Command {
	case p2p.RpcCommandBlock:
	case p2p.RpcCommandAddr:
	case p2p.RpcCommandTx:
	default:
		return nil
	}
	return nil
}

func (n *Node) Handshake(peer p2p.Peer) error {
	ourMsgVersion := p2p.MessageVersion{
		ProtocolVersion:   CurrentVersion,
		LatestBlockHeight: n.StateManager.GetLatestBlockHeight(),
	}
	msgBytes := ourMsgVersion.ToBytes()
	msg := p2p.NewMessage(msgBytes, p2p.RpcCommandVersion)

	msgEncoder := p2p.NewMessageEncoder(peer)
	err := msgEncoder.Encode(msg)
	if err != nil {
		return NewNodeError(fmt.Sprintf("failed to send version message: %s", err))
	}

	msgDecoder := p2p.NewMessageDecoder(peer)
	theirMsg, err := msgDecoder.Decode()
	if err != nil {
		return NewNodeError(fmt.Sprintf("failed to decode peer's version message: %s", err))
	}
	if theirMsg.Header.Command != p2p.RpcCommandVersion {
		return NewNodeError("peer handshake msg command is not version")
	}

	theirMsgVersion, err := theirMsg.VisitAsMessageVersion()
	if err != nil {
		return NewNodeError(fmt.Sprintf("failed to decode peer's version message payload: %s", err))
	}
	if theirMsgVersion.ProtocolVersion != CurrentVersion {
		return NewNodeError("peer protocol version is different from own")
	}

	if theirMsgVersion.LatestBlockHeight > n.StateManager.GetLatestBlockHeight() {
		// TODO: Send request for blocks to sync
	}

	log.Printf("handshake successful with peer %s", peer.GetAddr())
	return nil
}

type NodeError struct {
	Err string
}

func NewNodeError(e string) NodeError {
	return NodeError{Err: e}
}

func (e NodeError) Error() string {
	return fmt.Sprintf("Node error: %s", e.Err)
}
