package p2p

import (
	"bytes"
	"errors"
	"log"
	"net"
	"sync"
)

var NopHandshakeFunc HandshakeFunc = func(p Peer) error {
	return nil
}

type TCPPeer struct {
	// Implements the Peer interface
	net.Conn
}

func NewTCPPeer(conn net.Conn) *TCPPeer {
	return &TCPPeer{
		Conn: conn,
	}
}

func (t *TCPPeer) Send(data []byte) error {
	_, err := t.Write(data)
	return err
}

func (t *TCPPeer) GetAddr() string {
	return t.RemoteAddr().String()
}

type TCPTransportOpts struct {
	ListenAddr string
	SeedPeers  []string
	Handshake  HandshakeFunc
}

// Implements the Transport interface
type TCPTransport struct {
	TCPTransportOpts
	Listener net.Listener
	RpcCh    chan RPC

	PeerLock sync.RWMutex
	Peers    map[string]*TCPPeer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		RpcCh:            make(chan RPC, 1024),
		Peers:            make(map[string]*TCPPeer),
	}
}

func (t *TCPTransport) GetAddr() string {
	if t.Listener == nil {
		return ""
	} else {
		return t.Listener.Addr().String()
	}
}

func (t *TCPTransport) Close() error {
	return t.Listener.Close()
}

func (t *TCPTransport) Consume() <-chan RPC { // read-only channel of messages
	return t.RpcCh
}

func (t *TCPTransport) ListenAndServe() error {
	var err error
	t.Listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go t.StartAcceptLoop()
	log.Printf("TCP transport listening on port: %s", t.Listener.Addr().String())

	t.Bootstrap()
	return nil
}

func (t *TCPTransport) StartAcceptLoop() {
	for {
		conn, err := t.Listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			log.Printf("TCP accept error %v\n", err)
			continue
		}
		log.Printf("new incoming connection: [%s]\n", conn.RemoteAddr().String())

		go t.HandleConn(conn)
	}
}

func (t *TCPTransport) Connect(addr string) (net.Conn, error) {
	return net.Dial("tcp", addr)
}

func (t *TCPTransport) Bootstrap() {
	if t.SeedPeers == nil {
		return
	}
	for _, addr := range t.SeedPeers {
		t.HandleNewPeer(addr)
	}
}

func (t *TCPTransport) HandleConn(conn net.Conn) {
	peer := NewTCPPeer(conn)

	log.Println("trying to handshake")
	err := t.Handshake(peer)
	if err != nil {
		log.Printf("handshake with [%s] failed: %v", peer.GetAddr(), err)
		_ = peer.Close()
		return
	}
	log.Printf("handshake with [%s] successful", peer.GetAddr())

	t.AddPeer(peer)

	defer func() {
		log.Printf("dropping peer: [%s]", peer.GetAddr())
		t.RemovePeer(peer)
		_ = peer.Close()
	}()

	t.readLoop(peer)
}

func (t *TCPTransport) readLoop(peer *TCPPeer) {
	decoder := NewMessageDecoder(peer)
	for {
		msg, err := decoder.Decode()
		if err != nil {
			log.Printf("read error from [%s]: %v", peer.GetAddr(), err)
			break
		}
		t.RpcCh <- RPC{
			From:    peer.GetAddr(),
			Payload: msg.Payload,
			Command: msg.Header.Command,
		}
	}
}

func (t *TCPTransport) Broadcast(payload []byte, command RpcCommand) {
	msg := NewMessage(payload, command)
	var b bytes.Buffer
	encoder := NewMessageEncoder(&b)
	encoder.Encode(msg)
	t.PeerLock.RLock()
	defer t.PeerLock.RUnlock()
	for addr, peer := range t.Peers {
		err := peer.Send(b.Bytes())
		if err != nil {
			log.Printf("broadcast error: %v for addr [%s]", err, addr)
		}
	}
}

func (t *TCPTransport) HandleNewPeer(addr string) bool {
	_, ok := t.GetPeer(addr)
	if ok {
		return false
	}

	ownAddr := t.GetAddr()
	if addr == ownAddr {
		return false
	}

	conn, err := t.Connect(addr)
	if err != nil {
		log.Printf("couldn't connect to peer at addr [%s]", addr)
		return false
	}
	go t.HandleConn(conn)
	return true
}

func (t *TCPTransport) GetPeer(addr string) (*TCPPeer, bool) {
	t.PeerLock.RLock()
	defer t.PeerLock.RUnlock()
	peer, ok := t.Peers[addr]
	return peer, ok
}

func (t *TCPTransport) AddPeer(peer *TCPPeer) {
	t.PeerLock.Lock()
	defer t.PeerLock.Unlock()
	t.Peers[peer.GetAddr()] = peer
}

func (t *TCPTransport) RemovePeer(peer *TCPPeer) {
	t.PeerLock.Lock()
	defer t.PeerLock.Unlock()
	delete(t.Peers, peer.GetAddr())
}
