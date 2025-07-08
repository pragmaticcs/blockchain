package p2p

import (
	"bytes"
	"errors"
	"log"
	"net"
	"sync"
)

type Handshake func()

type TCPPeer struct {
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

type TCPTransportOpts struct {
	ListenAddr string
	SeedPeers  []string
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

func (t *TCPTransport) Close() error {
	return t.Listener.Close()
}

func (t *TCPTransport) Consume() <-chan RPC { // read-only channel of messages
	return t.RpcCh
}

func (t *TCPTransport) Start() error {
	var err error
	t.Listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go t.StartAcceptLoop()
	log.Printf("TCP transport listening on port: %s", t.ListenAddr)

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
		log.Printf("new incoming connection %s\n", conn.RemoteAddr().String())

		go t.HandleConn(conn)
	}
}

func (t *TCPTransport) Connect(addr string) (net.Conn, error) {
	return net.Dial("tcp", addr)
}

func (t *TCPTransport) Bootstrap() {
	for _, addr := range t.SeedPeers {
		conn, err := t.Connect(addr)
		if err != nil {
			// TODO: Handle periodic reconnection
			log.Printf("Couldn't connect to peer at addr [%s]", addr)
			continue
		}
		go t.HandleConn(conn)
	}
}

func (t *TCPTransport) HandleConn(conn net.Conn) {
	peer := NewTCPPeer(conn)
	t.AddPeer(peer)

	defer func() {
		log.Printf("dropping peer: %s", peer.RemoteAddr().String())
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
			log.Printf("read error from %s: %v", peer.RemoteAddr().String(), err)
			break
		}
		t.RpcCh <- RPC{
			From:    peer.RemoteAddr().String(),
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

func (t *TCPTransport) AddPeer(peer *TCPPeer) {
	t.PeerLock.Lock()
	defer t.PeerLock.Unlock()
	t.Peers[peer.RemoteAddr().String()] = peer
}

func (t *TCPTransport) RemovePeer(peer *TCPPeer) {
	t.PeerLock.Lock()
	defer t.PeerLock.Unlock()
	delete(t.Peers, peer.RemoteAddr().String())
}
