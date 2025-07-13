package node

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"slices"
	"sync"
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
	stateLock        sync.Mutex
	syncPeer         p2p.Peer
	inFlightRequests map[string]time.Time
	inFlightLock     sync.RWMutex
	// Channels
	quitCh        chan struct{}
	syncRequestCh chan SyncRequest // Maybe replace with RPC to self?
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
		stateLock:        sync.Mutex{},
		syncPeer:         nil,
		inFlightRequests: make(map[string]time.Time),
		inFlightLock:     sync.RWMutex{},
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
		return n.HandleRpcBlock(rpc)
	case p2p.RpcCommandAddr:
	case p2p.RpcCommandGetAddr:
	case p2p.RpcCommandTx:
	case p2p.RpcCommandGetBlocks:
		return n.HandleRpcGetBlocks(rpc)
	case p2p.RpcCommandGetData:
		return n.HandleRpcGetData(rpc)
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
		return NewHandshakeError(fmt.Sprintf("failed to send version message: %s", err))
	}

	// Wait for response
	msgDecoder := p2p.NewMessageDecoder(peer)
	theirMsg, err := msgDecoder.Decode()
	if err != nil {
		return NewHandshakeError(fmt.Sprintf("failed to decode peer's version message: %s", err))
	}
	if theirMsg.Header.Command != p2p.RpcCommandVersion {
		return NewHandshakeError("peer handshake msg command is not version")
	}

	theirMsgVersion, err := theirMsg.VisitAsMessageVersion()
	if err != nil {
		return NewHandshakeError(fmt.Sprintf("failed to decode peer's version message payload: %s", err))
	}
	if theirMsgVersion.ProtocolVersion != core.CurrentVersion {
		return NewHandshakeError("peer protocol version is different from own")
	}

	verack := p2p.NewMessage(nil, p2p.PrcCommandVerack)
	err = msgEncoder.Encode(verack)
	if err != nil {
		return NewHandshakeError(fmt.Sprintf("failed to send verack message: %s", err))
	}

	theirMsg, err = msgDecoder.Decode()
	if err != nil {
		return NewHandshakeError(fmt.Sprintf("failed to decode peer's verack message: %s", err))
	}
	if theirMsg.Header.Command != p2p.PrcCommandVerack {
		return NewHandshakeError("peer handshake response command is not verack")
	}

	if theirMsgVersion.LatestBlockHeight > n.StateManager.GetLatestBlockHeight() {
		n.sendSyncRequest(peer, theirMsgVersion.LatestBlockHeight)
	}

	return nil
}

func (n *Node) StartSync(peer p2p.Peer) error {
	n.setStateSync(peer)
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

func (n *Node) setStateIdle() {
	n.stateLock.Lock()
	defer n.stateLock.Unlock()
	n.state = NodeStateIdle
	n.syncPeer = nil
}

func (n *Node) setStateSync(peer p2p.Peer) {
	n.stateLock.Lock()
	defer n.stateLock.Unlock()
	n.state = NodeStateSync
	n.syncPeer = peer
}

func (n *Node) HandlePeriodicTasks() error {
	return nil
}

func (n *Node) tryFinishSync() {
	if n.noInFlightRequests() {
		log.Println("finished syncing")
		n.setStateIdle()
	}
}

func (n *Node) addInFlightRequests(items []p2p.InvVector) {
	n.inFlightLock.Lock()
	defer n.inFlightLock.Unlock()
	for _, item := range items {
		key := hex.EncodeToString(item.Hash)
		n.inFlightRequests[key] = time.Now()
	}
}

func (n *Node) isInFlightRequest(hash []byte) bool {
	n.inFlightLock.RLock()
	defer n.inFlightLock.RUnlock()
	key := hex.EncodeToString(hash)
	_, ok := n.inFlightRequests[key]
	return ok
}

func (n *Node) removeInFlightRequest(hash []byte) {
	n.inFlightLock.Lock()
	defer n.inFlightLock.Unlock()
	key := hex.EncodeToString(hash)
	delete(n.inFlightRequests, key)
}

func (n *Node) noInFlightRequests() bool {
	n.inFlightLock.RLock()
	defer n.inFlightLock.RUnlock()
	return len(n.inFlightRequests) == 0
}

func (n *Node) HandleRpcInvIdle(rpc p2p.RPC) error {
	msgInv, err := rpc.VisitAsMessageInv()
	if err != nil {
		return NewNodeError("sync error: couldn't decode message inv")
	}
	filteredItems := n.filterExistingItems(msgInv.Items)
	if len(filteredItems) == 0 {
		return nil
	}
	n.addInFlightRequests(filteredItems)
	msgGetData := p2p.MessageGetData{
		Items: filteredItems,
	}

	payload, err := msgGetData.ToBytes()
	if err != nil {
		return err
	}
	msg := p2p.NewMessage(payload, p2p.RpcCommandGetData)
	return n.SendMessage(rpc.From, msg)
}

func (n *Node) HandleRpcInvSync(rpc p2p.RPC) error {
	if rpc.From != n.syncPeer.GetAddr() {
		return nil
	}
	return n.HandleRpcInvIdle(rpc)
}

func (n *Node) filterExistingItems(items []p2p.InvVector) []p2p.InvVector {
	return slices.DeleteFunc(items, n.itemExists)
}

func (n *Node) itemExists(item p2p.InvVector) bool {
	// TODO: query items more efficiently, not one by one
	switch item.Type {
	case p2p.InvTypeBlock:
		return n.StateManager.BlockExists(item.Hash)
	case p2p.InvTypeTx:
		// TODO: implement mempool
		return false
	}
	return false
}

func (n *Node) HandleRpcInv(rpc p2p.RPC) error {
	switch n.state {
	case NodeStateIdle:
		return n.HandleRpcInvIdle(rpc)
	case NodeStateSync:
		return n.HandleRpcInvSync(rpc)
	}
	return nil
}

func (n *Node) HandleRpcGetData(rpc p2p.RPC) error {
	msgGetData, err := rpc.VisitAsMessageInv()
	if err != nil {
		return err
	}
	for _, item := range msgGetData.Items {
		switch item.Type {
		case p2p.InvTypeBlock:
			return n.HandleGetDataBlock(item, rpc.From)
		case p2p.InvTypeTx:
			return nil
		}
	}
	return nil
}

func (n *Node) HandleGetDataBlock(item p2p.InvVector, addr string) error {
	if !n.StateManager.BlockExists(item.Hash) {
		messageNotFound := p2p.NewMessage(nil, p2p.RpcCommandNotFound)
		return n.SendMessage(addr, messageNotFound)
	}

	// TODO: add a method to the state manager for handling fetching blocks
	block, err := n.StateManager.GetBlock(item.Hash)
	if err != nil {
		return err
	}
	msgBlock := p2p.MessageBlock{
		Block: *block,
	}
	payload, err := msgBlock.ToBytes()
	if err != nil {
		return err
	}
	msg := p2p.NewMessage(payload, p2p.RpcCommandBlock)
	return n.SendMessage(addr, msg)
}

func (n *Node) HandleRpcBlock(rpc p2p.RPC) error {
	if n.state == NodeStateSync && rpc.From != n.syncPeer.GetAddr() {
		return nil
	}

	msgBlock, err := rpc.VisitAsMessageBlock()
	if err != nil {
		return err
	}
	block := &msgBlock.Block
	if n.state == NodeStateSync && !n.isInFlightRequest(block.GetHash()) {
		return nil
	}
	valid := n.Validate(block)
	if !valid {
		return NewNodeError(fmt.Sprintf("block received from %s is not valid", rpc.From))
	}
	err = n.StateManager.Update(block)
	if err != nil {
		return err
	}
	if n.state == NodeStateSync {
		n.removeInFlightRequest(block.GetHash())
		n.tryFinishSync()
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

type HandshakeError struct {
	Err string
}

func NewHandshakeError(e string) HandshakeError {
	return HandshakeError{Err: e}
}

func (e HandshakeError) Error() string {
	return fmt.Sprintf("Handshake error: %s", e.Err)
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
