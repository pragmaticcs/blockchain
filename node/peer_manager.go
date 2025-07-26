package node

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pragmaticcs/blockchain/p2p"
)

const PeerTimeout time.Duration = 10 * time.Second

type PeerStatus uint8

const (
	PeerStatusHandshake PeerStatus = iota
	PeerStatusConnected
)

type PeerState struct {
	peer p2p.Peer

	Blockheight    uint32
	PeerStatus     PeerStatus
	Score          int
	LastSeenOnline time.Time
}

type PeerManager struct {
	node  *Node
	peers map[string]*PeerState
	lock  sync.RWMutex
}

func NewPeerManager(node *Node) *PeerManager {
	return &PeerManager{
		node:  node,
		peers: make(map[string]*PeerState),
		lock:  sync.RWMutex{},
	}
}

func (p *PeerManager) AddPeer(peer p2p.Peer) {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, exists := p.getPeerUnlocked(peer.GetAddr())
	if exists {
		return
	}

	peerState := &PeerState{
		peer:           peer,
		Blockheight:    1,
		Score:          0,
		PeerStatus:     PeerStatusHandshake,
		LastSeenOnline: time.Now(),
	}
	p.peers[peer.GetAddr()] = peerState
}

func (p *PeerManager) FinishHandshake(peer p2p.Peer, height uint32) {
	p.lock.Lock()
	defer p.lock.Unlock()

	state, ok := p.getPeerUnlocked(peer.GetAddr())
	if !ok {
		return
	}
	state.PeerStatus = PeerStatusConnected
	state.Blockheight = height
	state.LastSeenOnline = time.Now()
}

func (p *PeerManager) SendMessage(addr string, msg *p2p.Message) error {
	peerState, ok := p.GetPeer(addr)
	if !ok {
		return fmt.Errorf("%s is not a peer", addr)
	}
	if peerState.PeerStatus != PeerStatusConnected {
		return fmt.Errorf("peer %s is not connected", addr)
	}
	encoder := p2p.NewMessageEncoder(peerState.peer)
	return encoder.Encode(msg)
}

func (p *PeerManager) Broadcast(msg *p2p.Message) {
	syncPeers := p.GetBroadcastCandidates()

	var msgEncoder p2p.MessageEncoder
	for _, peer := range syncPeers {
		msgEncoder = *p2p.NewMessageEncoder(peer)
		err := msgEncoder.Encode(msg)
		if err != nil {
			log.Printf("broadcast error to peer [%s]: %s", peer.GetAddr(), err)
		}
	}
}

func (p *PeerManager) RemovePeer(peer p2p.Peer) {
	if peer == nil {
		return
	}
	p.lock.Lock()
	defer p.lock.Unlock()

	addr := peer.GetAddr()
	delete(p.peers, addr)
}

func (p *PeerManager) getPeerUnlocked(addr string) (*PeerState, bool) {
	// Without lock
	peerState, ok := p.peers[addr]
	return peerState, ok
}

func (p *PeerManager) GetPeer(addr string) (*PeerState, bool) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	peerState, ok := p.peers[addr]
	return peerState, ok
}

func (p *PeerManager) GetSyncCandidates(ownHeight uint32) []p2p.Peer {
	p.lock.RLock()
	defer p.lock.RUnlock()

	var candidatePeers []p2p.Peer

	for _, state := range p.peers {
		if state.PeerStatus == PeerStatusConnected && state.Blockheight > ownHeight {
			candidatePeers = append(candidatePeers, state.peer)
		}
	}
	return candidatePeers
}

func (p *PeerManager) GetBroadcastCandidates() []p2p.Peer {
	p.lock.RLock()
	defer p.lock.RUnlock()

	var candidatePeers []p2p.Peer

	for _, state := range p.peers {
		if state.PeerStatus == PeerStatusConnected {
			candidatePeers = append(candidatePeers, state.peer)
		}
	}
	return candidatePeers
}

func (p *PeerManager) FindSyncPeer(ownHeight uint32) (p2p.Peer, bool) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	var bestPeer p2p.Peer = nil
	var bestHeight uint32 = ownHeight
	for _, state := range p.peers {
		if state.PeerStatus == PeerStatusConnected && state.Blockheight > bestHeight {
			bestHeight = state.Blockheight
			bestPeer = state.peer
		}
	}
	if bestPeer == nil {
		return nil, false
	}
	return bestPeer, true
}

func (p *PeerManager) PenalizePeer(penalty int, peerAddr string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	state, ok := p.getPeerUnlocked(peerAddr)
	if !ok {
		return
	}
	state.Score -= penalty
}

func (p *PeerManager) RewardPeer(reward int, peerAddr string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	state, ok := p.getPeerUnlocked(peerAddr)
	if !ok {
		return
	}
	state.Score += reward
}

func (p *PeerManager) FindUnresponsivePeers() []p2p.Peer {
	p.lock.RLock()
	defer p.lock.RUnlock()

	var unsresponsivePeers []p2p.Peer

	for _, state := range p.peers {
		if state.PeerStatus == PeerStatusConnected &&
			state.LastSeenOnline.Add(PeerTimeout).Before(time.Now()) {
			unsresponsivePeers = append(unsresponsivePeers, state.peer)
		}
	}
	return unsresponsivePeers
}

func (p *PeerManager) PingPeers() {
	peersToPing := p.GetBroadcastCandidates()
	msg := p2p.NewMessage(nil, p2p.RpcCommandPing)

	for _, peer := range peersToPing {
		err := p.SendMessage(peer.GetAddr(), msg)
		if err != nil {
			log.Printf("failed to send ping to %s: %s", peer.GetAddr(), err)
		}
	}
}

func (p *PeerManager) TrimPeers() {
	unsresponsivePeers := p.FindUnresponsivePeers()
	for _, peer := range unsresponsivePeers {
		log.Printf("trimming unresponsive peer: %s", peer.GetAddr())
		p.DisconnectPeer(peer)
	}
}

func (p *PeerManager) DisconnectPeer(peer p2p.Peer) {
	p.RemovePeer(peer)
	_ = peer.Close()
}

func (p *PeerManager) MarkSeenOnline(addr string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	state, ok := p.getPeerUnlocked(addr)
	if !ok {
		return
	}
	state.LastSeenOnline = time.Now()
}

func (p *PeerManager) PeerDisconnectFunc(peer p2p.Peer) {
	log.Printf("removing peer %s\n", peer.GetAddr())
	p.RemovePeer(peer)
}
