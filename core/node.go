package core

import (
	"fmt"
	"log"

	"github.com/pragmaticcs/blockchain/p2p"
	"github.com/pragmaticcs/blockchain/wallet"
)

type NodeOpts struct {
	Port                  string
	SeedNodes             []string
	WalletFileDir         string
	BlockchainDatabaseDir string
}

func DefaultNodeOpts() NodeOpts {
	return NodeOpts{
		Port:                  "3000",
		SeedNodes:             nil,
		WalletFileDir:         "", // Save to current directory
		BlockchainDatabaseDir: "badger",
	}
}

type Node struct {
	StateManager *StateManager
	Mempool      *Mempool
	Transport    *p2p.TCPTransport // TODO: Replace with Transport interface

	// Channels
	quit chan struct{}
}

func NewNode(opts NodeOpts) (*Node, error) {
	var node Node

	transportOpts := p2p.TCPTransportOpts{
		ListenAddr: opts.Port,
		SeedPeers:  opts.SeedNodes,
		Handshake:  node.Handshake,
	}
	transport := p2p.NewTCPTransport(transportOpts)
	store, err := NewBadgerStore(opts.BlockchainDatabaseDir)
	if err != nil {
		return nil, NewNodeError(fmt.Sprintf("couldn't connect to the database: %s", err))
	}
	blockchain, err := InitBlockchain(store)
	if err != nil {
		return nil, NewNodeError(fmt.Sprintf("couldn't initialize the blockchain: %s", err))
	}
	wallets, err := wallet.InitWallets(opts.WalletFileDir)
	if err != nil {
		return nil, NewNodeError(fmt.Sprintf("couldn't initialize wallets: %s", err))
	}
	stateManager := NewStateManager(blockchain, wallets)

	node.Transport = transport
	node.StateManager = stateManager
	node.quit = make(chan struct{})
	node.Mempool = nil
	return &node, err
}

func NewNodeTesting(opts NodeOpts, b *Blockchain, w *wallet.Wallets) *Node {
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
	err := n.Transport.Start()
	if err != nil {
		return err
	}

	return n.MainLoop()
}

func (n *Node) Stop() {
	// Handles graceful shutdown
	_ = n.Transport.Close()
	_ = n.StateManager.Close()

}

func (n *Node) sendQuitSignal() {
	// Sends a quit signal
	n.quit <- struct{}{}
}

func (n *Node) MainLoop() error {
	for {
		select {
		case rpc := <-n.Transport.RpcCh:
			err := n.HandleRpc(rpc)
			if err != nil {
				return err
			}
		case <-n.quit:
			return nil
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

	// log.Printf("Message version from peer %s: %+v\n", peer.GetAddr(), theirMsgVersion)

	if theirMsgVersion.LatestBlockHeight > n.StateManager.GetLatestBlockHeight() {
		// TODO: Send request for blocks to sync
		n.Sync(peer)
	}

	log.Printf("handshake successful with peer %s", peer.GetAddr())
	return nil
}

func (n *Node) Sync(peer p2p.Peer) error {
	log.Println("syncing with peer")
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
