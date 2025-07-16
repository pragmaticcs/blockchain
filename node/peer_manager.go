package node

import (
	"fmt"
	"sync"
	"time"

	"github.com/pragmaticcs/blockchain/p2p"
)

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
	peers map[string]*PeerState
	lock  sync.RWMutex
}

func NewPeerManager() *PeerManager {
	return &PeerManager{
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

func (p *PeerManager) RemovePeer(peer p2p.Peer) {
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

	return p.getPeerUnlocked(addr)
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

func (p *PeerManager) PenalizePeer(penalty int, peer p2p.Peer) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	state, ok := p.getPeerUnlocked(peer.GetAddr())
	if !ok {
		return
	}
	state.Score -= penalty
}

func (p *PeerManager) RewardPeer(reward int, peer p2p.Peer) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	state, ok := p.getPeerUnlocked(peer.GetAddr())
	if !ok {
		return
	}
	state.Score += reward
}
