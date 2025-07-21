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

var (
	ErrInvalidBlock = NewNodeError("block is not valid")
	ErrInvalidTx    = NewNodeError("transaction is not valid")
)

// Maximum number of goroutines handling different
// job in the mainloop of the node
const MaxWorkers int = 8
const DefaultMineReward = 5

type PeriodicTask func() error

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
	Mempool      *Mempool
	Transport    *p2p.TCPTransport // TODO: Replace with Transport interface
	SyncManager  *SyncManager
	PeerManager  *PeerManager
	Miner        *Miner

	// Keep track of current requests
	inFlightRequests map[string]time.Time
	inFlightLock     sync.RWMutex
	// Channels
	rpcCh  chan p2p.RPC
	quitCh chan struct{}
	syncCh chan struct{}
	ticker *time.Ticker
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
	stateManager, err := core.NewStateManager(blockchain, wallet)
	if err != nil {
		return nil, NewNodeError(fmt.Sprintf("couldn't initialize the state manager: %s", err))
	}
	node := &Node{
		StateManager:     stateManager,
		Mempool:          NewMempool(),
		Transport:        nil,
		PeerManager:      NewPeerManager(),
		SyncManager:      nil,
		Miner:            nil,
		inFlightRequests: make(map[string]time.Time),
		inFlightLock:     sync.RWMutex{},
		rpcCh:            make(chan p2p.RPC, 100),
		quitCh:           make(chan struct{}),
		syncCh:           make(chan struct{}),
		ticker:           time.NewTicker(5 * time.Second),
	}
	transportOpts := p2p.TCPTransportOpts{
		ListenAddr: opts.Port,
		SeedPeers:  opts.SeedNodes,
		Handshake:  node.Handshake,
	}
	node.Transport = p2p.NewTCPTransport(transportOpts)
	node.SyncManager = NewSyncManager(node.PeerManager, node)
	node.Miner = NewMiner(node)
	return node, err
}

func (n *Node) launchWorkerPool(jobs int) {
	for range jobs {
		go n.rpcWorker()
	}
}

func (n *Node) rpcWorker() {
	// RPC worker listens on the channels of rpcs
	for rpc := range n.rpcCh {
		err := n.HandleRpc(rpc)
		if err != nil {
			log.Printf("error handling RPC from %s: %v", rpc.From, err)
		}
	}
}

func (n *Node) Start() error {
	n.launchWorkerPool(MaxWorkers)
	n.Miner.Start()
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
	close(n.quitCh)
	close(n.rpcCh)
	n.ticker.Stop()
	_ = n.Transport.Close()
	_ = n.StateManager.Close()
}

func (n *Node) sendQuitSignal() {
	// Sends a quit signal
	n.quitCh <- struct{}{}
}

func (n *Node) sendSyncSignal() {
	// Sends a sync signal
	n.syncCh <- struct{}{}
}

func (n *Node) MainLoop() {
	for {
		select {
		case rpc := <-n.Transport.Consume():
			// Redirect rpc to workerpool
			n.rpcCh <- rpc
		case <-n.syncCh:
			go n.SyncManager.attemptSync()
		case <-n.ticker.C:
			go n.HandlePeriodicTasks()
		case <-n.quitCh:
			log.Println("MainLoop: received quit signal, shutting down.")
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
		return n.HandleRpcAddr(rpc)
	case p2p.RpcCommandGetAddr:
		return n.HandleRpcGetAddr(rpc)
	case p2p.RpcCommandTx:
		return n.HandleRpcTx(rpc)
	case p2p.RpcCommandGetBlocks:
		return n.HandleRpcGetBlocks(rpc)
	case p2p.RpcCommandGetData:
		return n.HandleRpcGetData(rpc)
	case p2p.RpcCommandNotFound:
		return n.HandleRpcNotFound(rpc)
	default:
		return nil
	}
}

func (n *Node) Handshake(peer p2p.Peer) error {
	n.PeerManager.AddPeer(peer)
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

	n.PeerManager.FinishHandshake(peer, theirMsgVersion.LatestBlockHeight)
	n.RequestAddr(peer)

	if theirMsgVersion.LatestBlockHeight > n.StateManager.GetLatestBlockHeight() {
		n.sendSyncSignal()
	}

	return nil
}

func (n *Node) HandlePeriodicTasks() error {
	peers := n.PeerManager.GetBroadcastCandidates()
	log.Printf("connected to %d peers", len(peers))
	return nil
}

func (n *Node) RequestAddr(peer p2p.Peer) error {
	msg := p2p.NewMessage(nil, p2p.RpcCommandGetAddr)
	return n.PeerManager.SendMessage(peer.GetAddr(), msg)
}

func (n *Node) HandleRpcNotFound(rpc p2p.RPC) error {
	// TODO: Handle Not Found response here and in sync manager
	log.Printf("NOT FOUND msg from %s", rpc.From)
	return nil
}

func (n *Node) HandleRpcGetAddr(rpc p2p.RPC) error {
	peers := n.PeerManager.GetBroadcastCandidates()
	var peerAddrs []string
	for _, peer := range peers {
		if peer.GetAddr() != rpc.From {
			peerAddrs = append(peerAddrs, peer.GetAddr())
		}
	}
	msgAddr := p2p.MessageAddr{
		AddrList: peerAddrs,
	}
	payload, _ := msgAddr.ToBytes()
	msg := p2p.NewMessage(payload, p2p.RpcCommandAddr)
	return n.PeerManager.SendMessage(rpc.From, msg)
}

func (n *Node) HandleRpcAddr(rpc p2p.RPC) error {
	msgAddr, err := rpc.VisitAsMessageAddr()
	if err != nil {
		return NewNodeError("couldn't decode message addr")
	}
	for _, addr := range msgAddr.AddrList {
		log.Printf("got new peer addr %s", addr)
		n.Transport.HandleNewPeer(addr)
	}
	return nil
}

func (n *Node) HandleRpcInv(rpc p2p.RPC) error {
	if n.SyncManager.IsSyncing() {
		// Delegate the task to Sync Manager if syncing
		return n.SyncManager.HandleRpcInv(rpc)
	}
	msgInv, err := rpc.VisitAsMessageInv()
	if err != nil {
		return NewNodeError("couldn't decode message inv")
	}
	filteredItems := n.filterExistingItems(msgInv.Items)
	if len(filteredItems) == 0 {
		return nil
	}
	msgGetData := p2p.MessageGetData{
		Items: filteredItems,
	}

	payload, err := msgGetData.ToBytes()
	if err != nil {
		return err
	}
	msg := p2p.NewMessage(payload, p2p.RpcCommandGetData)
	return n.PeerManager.SendMessage(rpc.From, msg)
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
		return n.Mempool.TxExists(item.Hash)
	}
	return false
}

func (n *Node) HandleRpcGetData(rpc p2p.RPC) error {
	msgGetData, err := rpc.VisitAsMessageInv()
	if err != nil {
		return err
	}
	for _, item := range msgGetData.Items {
		switch item.Type {
		case p2p.InvTypeBlock:
			err := n.HandleGetDataBlock(item, rpc.From)
			if err != nil {
				return err
			}
		case p2p.InvTypeTx:
			err := n.HandleGetDataTx(item, rpc.From)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *Node) HandleGetDataTx(item p2p.InvVector, addr string) error {
	tx, ok := n.Mempool.GetTransaction(core.TransactionID(item.Hash))

	if !ok {
		messageNotFound := p2p.NewMessage(nil, p2p.RpcCommandNotFound)
		return n.PeerManager.SendMessage(addr, messageNotFound)
	}

	msgTx := p2p.MessageTx{
		Tx: *tx,
	}
	payload, err := msgTx.ToBytes()
	if err != nil {
		return err
	}
	msg := p2p.NewMessage(payload, p2p.RpcCommandTx)
	return n.PeerManager.SendMessage(addr, msg)
}

func (n *Node) HandleGetDataBlock(item p2p.InvVector, addr string) error {
	if !n.StateManager.BlockExists(item.Hash) {
		messageNotFound := p2p.NewMessage(nil, p2p.RpcCommandNotFound)
		return n.PeerManager.SendMessage(addr, messageNotFound)
	}

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
	return n.PeerManager.SendMessage(addr, msg)
}

func (n *Node) HandleRpcTx(rpc p2p.RPC) error {
	msgTx, err := rpc.VisitAsMessageTx()
	if err != nil {
		return err
	}

	if n.Mempool.TxExists(msgTx.Tx.ID[:]) {
		return nil
	}

	valid, _ := n.StateManager.Blockchain.VerifyTransaction(&msgTx.Tx)
	if !valid {
		// Penalize peer for invalid tx
		n.PeerManager.PenalizePeer(20, rpc.From)
		return ErrInvalidTx
	}

	n.Mempool.AddTx(&msgTx.Tx)
	n.Miner.StartMining()

	n.BroadcastInv(msgTx.Tx.ID[:], p2p.InvTypeTx)
	return nil
}

func (n *Node) HandleRpcBlock(rpc p2p.RPC) error {
	if n.SyncManager.IsSyncing() {
		// Delegate the task to Sync Manager
		return n.SyncManager.HandleRpcBlock(rpc)
	}

	msgBlock, err := rpc.VisitAsMessageBlock()
	if err != nil {
		return err
	}
	block := &msgBlock.Block

	err = n.ProcessBlock(block)
	if err == ErrInvalidBlock {
		n.PeerManager.PenalizePeer(30, rpc.From)
		return nil
	}
	// Stop mining and remove txs from the mempool that have been mined
	n.Miner.Reset()
	n.Mempool.Clear(block)

	// Broadcast new block to peers
	n.BroadcastInv(block.GetHash(), p2p.InvTypeBlock)
	return err
}

func (n *Node) ProcessBlock(block *core.Block) error {
	valid, err := n.StateManager.Validate(block)
	if err != nil {
		return err
	}
	if !valid {
		return ErrInvalidBlock
	}
	log.Printf("processed block %s", hex.EncodeToString(block.GetHash()))

	return n.StateManager.Update(block)
}

func (n *Node) HandleRpcGetBlocks(rpc p2p.RPC) error {
	msgGetBlocks, err := rpc.VisitAsMessageGetBlocks()
	if err != nil {
		return err
	}
	if !n.StateManager.Blockchain.Store.Exists(msgGetBlocks.From) {
		messageNotFound := p2p.NewMessage(nil, p2p.RpcCommandNotFound)
		return n.PeerManager.SendMessage(rpc.From, messageNotFound)
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
	slices.Reverse(items)
	messageInv := p2p.MessageInv{
		Items: items,
	}
	payload, err := messageInv.ToBytes()
	if err != nil {
		return err
	}
	msg := p2p.NewMessage(payload, p2p.RpcCommandInv)
	return n.PeerManager.SendMessage(rpc.From, msg)
}

func (n *Node) BroadcastInv(hash []byte, invType p2p.InvType) {
	msgInv := p2p.MessageInv{
		Items: []p2p.InvVector{
			{
				Type: p2p.InvTypeBlock,
				Hash: hash,
			},
		},
	}
	payload, _ := msgInv.ToBytes()
	msg := p2p.NewMessage(payload, p2p.RpcCommandInv)
	n.PeerManager.Broadcast(msg)
}

func (n *Node) RequestBlocks(peer p2p.Peer) error {
	log.Printf("attempting syncing with peer %s", peer.GetAddr())

	// Request blocks from our blockchain head
	msgGetBlocks := p2p.MessageGetBlocks{
		From: n.StateManager.GetBlockchainHead(),
	}
	msgPayload, err := msgGetBlocks.ToBytes()
	if err != nil {
		return err
	}
	msg := p2p.NewMessage(msgPayload, p2p.RpcCommandGetBlocks)

	return n.PeerManager.SendMessage(peer.GetAddr(), msg)
}

func (n *Node) RequestItems(peer p2p.Peer, items []p2p.InvVector) error {
	log.Printf("requesting items from peer %s", peer.GetAddr())

	msgGetData := p2p.MessageGetData{
		Items: items,
	}
	msgPayload, err := msgGetData.ToBytes()
	if err != nil {
		return err
	}
	msg := p2p.NewMessage(msgPayload, p2p.RpcCommandGetData)

	return n.PeerManager.SendMessage(peer.GetAddr(), msg)
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
