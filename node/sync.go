package node

import (
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/pragmaticcs/blockchain/core"
	"github.com/pragmaticcs/blockchain/p2p"
)

const MaxItemsPerJob int = 16

type DownloadJob struct {
	Items      []p2p.InvVector
	AssignedAt time.Time
	AssignedTo p2p.Peer
}

// Node will delegate the task of syncing to this struct
type SyncManager struct {
	PeerManager *PeerManager
	Node        *Node

	pendingJobs      chan *DownloadJob
	inFlightRequests map[string]*DownloadJob
	// This map is to ensure processing blocks in correct order
	pendingBlocks map[uint32]*core.Block // key is the blockheight

	syncing          bool     // State to check if we are syncing
	invPeer          p2p.Peer // Peer that we got our block hashes from
	lock             sync.RWMutex
	pendingBlockLock sync.RWMutex
}

func NewSyncManager(pm *PeerManager, n *Node) *SyncManager {
	return &SyncManager{
		PeerManager: pm,
		Node:        n,

		pendingJobs:      make(chan *DownloadJob),
		inFlightRequests: make(map[string]*DownloadJob),
		pendingBlocks:    make(map[uint32]*core.Block),
		syncing:          false,
		invPeer:          nil,
		lock:             sync.RWMutex{},
		pendingBlockLock: sync.RWMutex{},
	}
}

func (s *SyncManager) attemptSync() error {
	s.lock.Lock()
	if s.syncing {
		s.lock.Unlock()
		return nil
	}

	ownHeight := s.Node.StateManager.GetLatestBlockHeight()
	invPeer, ok := s.PeerManager.FindSyncPeer(ownHeight)
	if !ok {
		s.lock.Unlock()
		log.Println("couldn't find a peer to sync inventories with")
		return nil
	}

	log.Printf("syncing inventories with %s", invPeer.GetAddr())
	s.syncing = true
	s.invPeer = invPeer
	s.lock.Unlock()

	go s.assignJobLoop(ownHeight)
	return s.Node.RequestBlocks(invPeer)
}

func (s *SyncManager) assignJobLoop(ownHeight uint32) error {
	log.Println("assign job loop started")
	for job := range s.pendingJobs {
		var assignedPeer p2p.Peer
		for {
			syncCandidates := s.PeerManager.GetSyncCandidates(ownHeight)
			if len(syncCandidates) == 0 {
				log.Println("couldn't find peer to assign job to, waiting")
				time.Sleep(time.Second)
				continue
			}
			// TODO: Improve assigning peers
			assignedPeer = syncCandidates[0]
			break
		}

		err := s.assignJobToPeer(job, assignedPeer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SyncManager) assignJobToPeer(job *DownloadJob, peer p2p.Peer) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Printf("assigning new job to peer %s", peer.GetAddr())

	job.AssignedAt = time.Now()
	job.AssignedTo = peer
	s.addInFlightRequests(job.Items, job)
	s.Node.RequestItems(peer, job.Items)
	return nil
}

func (s *SyncManager) HandleRpcInv(rpc p2p.RPC) error {
	s.lock.Lock()

	// Only accept inventory from our syncPeer
	// TODO: Handle timeouts for the sync peer if he doesn't respond
	// and reassign the syncPeer to another peer
	if !s.syncing || rpc.From != s.invPeer.GetAddr() {
		return nil
	}
	s.lock.Unlock()

	msgInv, err := rpc.VisitAsMessageInv()
	if err != nil {
		return err
	}

	log.Printf("received inventory of %d items from %s", len(msgInv.Items), rpc.From)

	s.QueuePendingJobs(msgInv.Items)
	return nil
}

func (s *SyncManager) QueuePendingJobs(items []p2p.InvVector) {
	itemCount := len(items)
	for i := 0; i < itemCount; i += MaxItemsPerJob {
		end := min(i+MaxItemsPerJob, itemCount)
		jobItems := items[i:end]

		job := &DownloadJob{
			Items: jobItems,
		}
		s.pendingJobs <- job
	}
}

func (s *SyncManager) IsSyncing() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.syncing
}

func (s *SyncManager) HandleRpcBlock(rpc p2p.RPC) error {
	log.Printf("recieved block from %s", rpc.From)
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.syncing {
		return nil
	}
	msgBlock, err := rpc.VisitAsMessageBlock()
	if err != nil {
		return nil
	}
	blockHash := msgBlock.Block.GetHash()
	if !s.isInFlightRequest(blockHash) {
		return nil
	}

	block := msgBlock.Block
	err = s.addPendingBlock(&block)
	if err != nil {
		return err
	}

	s.removeInFlightRequest(blockHash)

	// If there are other requests -> return
	if s.noInFlightRequests() {
		s.finishSync()
		log.Println("finished syncing")
		// TODO: Check if other peers are ahead after syncing
	}
	return nil
}

func (s *SyncManager) addPendingBlock(block *core.Block) error {
	s.pendingBlockLock.Lock()
	defer s.pendingBlockLock.Unlock()

	s.pendingBlocks[block.Height] = block
	return s.processPendingBlocks()
}

func (s *SyncManager) processPendingBlocks() error {
	for {
		currHeight := s.Node.StateManager.GetLatestBlockHeight()
		block, ok := s.pendingBlocks[currHeight+1]
		if !ok {
			break
		}
		err := s.Node.ProcessBlock(block)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SyncManager) finishSync() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.syncing = false
	s.invPeer = nil
}

func (s *SyncManager) addInFlightRequests(items []p2p.InvVector, job *DownloadJob) {
	for _, item := range items {
		key := hex.EncodeToString(item.Hash)
		s.inFlightRequests[key] = job
	}
}

func (s *SyncManager) isInFlightRequest(hash []byte) bool {
	key := hex.EncodeToString(hash)
	_, ok := s.inFlightRequests[key]
	return ok
}

func (s *SyncManager) removeInFlightRequest(hash []byte) {
	key := hex.EncodeToString(hash)
	delete(s.inFlightRequests, key)
}

func (s *SyncManager) noInFlightRequests() bool {
	return len(s.inFlightRequests) == 0
}
