package node

import (
	"bytes"
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
	quitCh        chan struct{}
	syncRequestCh chan SyncRequest
}

func NewNode(opts NodeOpts) (*Node, error) {

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
	node := &Node{
		StateManager:     stateManager,
		Mempool:          nil,
		Transport:        nil,
		state:            NodeStateIdle,
		inFlightRequests: make(map[string]time.Time),
		quitCh:           make(chan struct{}),
		syncRequestCh:    make(chan SyncRequest),
	}
	transportOpts := p2p.TCPTransportOpts{
		ListenAddr: opts.Port,
		SeedPeers:  opts.SeedNodes,
		Handshake:  node.Handshake,
	}
	node.Transport = p2p.NewTCPTransport(transportOpts)
	return node, err
}

func (n *Node) Start() error {
	go n.MainLoop()
	err := n.Transport.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) Stop() {
	// Handles graceful shutdown
	n.sendQuitSignal()
	_ = n.Transport.Close()
	_ = n.StateManager.Close()
}

func (n *Node) sendQuitSignal() {
	// Sends a quit signal
	n.quitCh <- struct{}{}
}

func (n *Node) sendSyncRequest(peer p2p.Peer, blockheight uint32) {
	time.Sleep(time.Second)
	n.syncRequestCh <- SyncRequest{
		Peer:           peer,
		NewBlockHeight: blockheight,
	}
}

func (n *Node) MainLoop() {
	for {
		select {
		case rpc := <-n.Transport.RpcCh:
			go func(rpc p2p.RPC) {
				err := n.HandleRpc(rpc)
				if err != nil {
					log.Println(err)
				}
			}(rpc)
		case req := <-n.syncRequestCh:
			if n.state == NodeStateIdle {
				n.StartSync(req.Peer)
			}
		case <-n.quitCh:
			return
		}
	}
}

func (n *Node) HandleRpc(rpc p2p.RPC) error {
	switch rpc.Command {
	case p2p.RpcCommandInv:
		return n.HandleRpcInv(rpc)
	case p2p.RpcCommandBlock:
		// return n.HandleRpcBlock(rpc)
	case p2p.RpcCommandAddr:
	case p2p.RpcCommandGetAddr:
	case p2p.RpcCommandTx:
	case p2p.RpcCommandGetBlocks:
		return n.HandleRpcGetBlocks(rpc)
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
	msgBytes, _ := ourMsgVersion.ToBytes()
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

	// TODO: Add send a verack message

	if theirMsgVersion.LatestBlockHeight > n.StateManager.GetLatestBlockHeight() {
		n.sendSyncRequest(peer, theirMsgVersion.LatestBlockHeight)
	}

	return nil
}

func (n *Node) StartSync(peer p2p.Peer) error {
	n.state = NodeStateSync
	log.Println("attempting syncing with peer")

	// Request blocks our blockchain head
	msgGetBlocks := p2p.MessageGetBlocks{
		From: n.StateManager.Blockchain.Head,
	}
	msgPayload, err := msgGetBlocks.ToBytes()
	if err != nil {
		return err
	}
	msg := p2p.NewMessage(msgPayload, p2p.RpcCommandGetBlocks)

	return n.SendMessage(peer.GetAddr(), msg)
}

func (n *Node) HandlePeriodicTasks() error {
	return nil
}

func (n *Node) addInFlightRequests(items []p2p.InvVector) {
	for _, item := range items {
		n.inFlightRequests[item.ToMapKey()] = time.Now()
	}
}

func (n *Node) HandleRpcInv(rpc p2p.RPC) error {
	msgInv, err := rpc.VisitAsMessageInv()
	if err != nil {
		return NewNodeError("sync error: couldn't decode message inv")
	}
	for item := range msgInv.Items {
		fmt.Printf("item: %v\n", item)
	}
	return nil
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

func (n *Node) SendMessage(addr string, msg *p2p.Message) error {
	peer, ok := n.Transport.GetPeer(addr)
	if !ok {
		return NewNodeError("peer doesn't exist")
	}
	encoder := p2p.NewMessageEncoder(peer)
	return encoder.Encode(msg)
}

func (n *Node) HandleRpcGetBlocks(rpc p2p.RPC) error {
	msgGetBlocks, err := rpc.VisitAsMessageGetBlocks()
	if err != nil {
		return err
	}
	if !n.StateManager.Blockchain.Store.Exists(msgGetBlocks.From) {
		messageNotFound := p2p.NewMessage(nil, p2p.RpcCommandNotFound)
		return n.SendMessage(rpc.From, messageNotFound)
	}
	var items []p2p.InvVector
	for block := range n.StateManager.Blockchain.Iterate() {
		blockHash := block.GetHash()
		if bytes.Equal(blockHash, msgGetBlocks.From) {
			break
		}
		items = append(items, p2p.InvVector{
			Type: p2p.InvTypeBlock,
			Hash: blockHash,
		})
	}
	messageInv := p2p.MessageInv{
		Items: items,
	}

	payload, err := messageInv.ToBytes()
	if err != nil {
		return err
	}
	msg := p2p.NewMessage(payload, p2p.RpcCommandInv)
	return n.SendMessage(rpc.From, msg)
}

type SyncRequest struct {
	Peer           p2p.Peer
	NewBlockHeight uint32
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
