package node

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pragmaticcs/blockchain/core"
	"github.com/pragmaticcs/blockchain/p2p"
	"github.com/pragmaticcs/blockchain/wallet"
)

type NodeState uint8

const (
	NodeStateIdle NodeState = iota
	NodeStateSync
)

type NodeOpts struct {
	Port                  string
	SeedNodes             []string
	WalletDir             string
	WalletName            string
	BlockchainDatabaseDir string
	GenesisConfigDir      string
}

func DefaultNodeOpts() NodeOpts {
	dir, _ := os.Getwd()
	return NodeOpts{
		Port:                  "3000",
		SeedNodes:             nil,
		WalletName:            "",
		WalletDir:             dir,
		BlockchainDatabaseDir: "badger",
		GenesisConfigDir:      dir,
	}
}

type Node struct {
	StateManager *core.StateManager
	Mempool      *core.Mempool
	Transport    *p2p.TCPTransport // TODO: Replace with Transport interface

	// Keep track of current requests
	state            NodeState
	inFlightRequests map[string]time.Time
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
	store, err := core.NewBadgerStore(opts.BlockchainDatabaseDir)
	if err != nil {
		return nil, NewNodeError(fmt.Sprintf("couldn't connect to the database: %s", err))
	}
	blockchain, err := core.InitBlockchain(store, opts.GenesisConfigDir)
	if err != nil {
		return nil, NewNodeError(fmt.Sprintf("couldn't initialize the blockchain: %s", err))
	}
	wallet, err := wallet.LoadWalletFromFile(opts.WalletDir, opts.WalletName)

	if err != nil {
		return nil, NewNodeError(fmt.Sprintf("couldn't load wallet: %s", err))
	}
	stateManager := core.NewStateManager(blockchain, wallet)

	node.Transport = transport
	node.StateManager = stateManager
	node.quit = make(chan struct{})
	node.Mempool = nil
	node.inFlightRequests = make(map[string]time.Time)
	node.state = NodeStateIdle
	return &node, err
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
		return n.HandleRpcBlock(rpc)
	case p2p.RpcCommandAddr:
	case p2p.RpcCommandTx:
	case p2p.RpcCommandGetBlocks:
	case p2p.RpcCommandGetData:
	default:
		return nil
	}
	return nil
}

func (n *Node) Handshake(peer p2p.Peer) error {
	// Send a version message
	ourMsgVersion := p2p.MessageVersion{
		ProtocolVersion:   core.CurrentVersion,
		LatestBlockHeight: n.StateManager.GetLatestBlockHeight(),
	}
	msgBytes := ourMsgVersion.ToBytes()
	msg := p2p.NewMessage(msgBytes, p2p.RpcCommandVersion)

	msgEncoder := p2p.NewMessageEncoder(peer)
	err := msgEncoder.Encode(msg)
	if err != nil {
		return NewNodeError(fmt.Sprintf("handshake error: failed to send version message: %s", err))
	}

	// Wait for response
	msgDecoder := p2p.NewMessageDecoder(peer)
	theirMsg, err := msgDecoder.Decode()
	if err != nil {
		return NewNodeError(fmt.Sprintf("handshake error: failed to decode peer's version message: %s", err))
	}
	if theirMsg.Header.Command != p2p.RpcCommandVersion {
		return NewNodeError("handshake error: peer handshake msg command is not version")
	}

	theirMsgVersion, err := theirMsg.VisitAsMessageVersion()
	if err != nil {
		return NewNodeError(fmt.Sprintf("handshake error: failed to decode peer's version message payload: %s", err))
	}
	if theirMsgVersion.ProtocolVersion != core.CurrentVersion {
		return NewNodeError("handshake error: peer protocol version is different from own")
	}

	// log.Printf("Message version from peer %s: %+v\n", peer.GetAddr(), theirMsgVersion)

	if theirMsgVersion.LatestBlockHeight > n.StateManager.GetLatestBlockHeight() {
		n.StartSync(peer)
	}

	// TODO: Add send a verack message

	log.Printf("handshake successful with peer %s", peer.GetAddr())
	return nil
}

func (n *Node) StartSync(peer p2p.Peer) error {
	log.Println("attempting syncing with peer")

	// Request blocks our blockchain head
	msgGetBlocks := p2p.MessageGetBlocks{
		From: n.StateManager.Blockchain.Head,
	}
	msgPayload := msgGetBlocks.ToBytes()
	msg := p2p.NewMessage(msgPayload, p2p.RpcCommandGetBlocks)

	encoder := p2p.NewMessageEncoder(peer)
	err := encoder.Encode(msg)
	if err != nil {
		return NewNodeError(fmt.Sprintf("sync error: failed to get blocks message: %s", err))
	}

	// Wait for response

	decoder := p2p.NewMessageDecoder(peer)
	response, err := decoder.Decode()
	if err != nil {
		return NewNodeError(fmt.Sprintf("sync error: failed to decode peer's inv message: %s", err))
	}

	if response.Header.Command != p2p.RpcCommandInv {
		return NewNodeError("sync error: peers response to msg getblocks is not an inv message")
	}

	msgInv, err := response.VisitAsMessageInv()
	if err != nil {
		return NewNodeError(fmt.Sprintf("sync error: failed to decode peer's version message payload: %s", err))
	}
	items := msgInv.Items
	return n.SimpleSync(peer, items)
}

func (n *Node) SimpleSync(peer p2p.Peer, items []p2p.InvVector) error {
	// The non-parallel version of sync
	log.Println("requesting blocks from a single peer")

	// Request items
	msgGetData := p2p.MessageGetData{
		Items: items,
	}
	payload := msgGetData.ToBytes()
	msg := p2p.NewMessage(payload, p2p.RpcCommandGetData)

	encoder := p2p.NewMessageEncoder(peer)
	err := encoder.Encode(msg)

	if err != nil {
		return NewNodeError(fmt.Sprintf("sync error: failed to send get data message: %s", err))
	}
	return nil
}

func (n *Node) HandlePeriodicTasks() error {
	return nil
}

func (n *Node) addInFlightRequests(items []p2p.InvVector) {
	for _, item := range items {
		n.inFlightRequests[item.ToMapKey()] = time.Now()
	}
}

func (n *Node) HandleRpcBlock(rpc p2p.RPC) error {
	msgBlock, err := rpc.VisitAsMessageBlock()
	if err != nil {
		return err
	}
	block := &msgBlock.Block
	valid := n.Validate(block)
	if !valid {
		return NewNodeError(fmt.Sprintf("block received from %s is not valid", rpc.From))
	}
	err = n.StateManager.Update(block)
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) Validate(block *core.Block) bool {
	return true
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
