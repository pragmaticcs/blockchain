package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// Basically just a connection + other stuff
type TCPPeer struct {
	net.Conn
	wg *sync.WaitGroup
}

func (p *TCPPeer) Send(data []byte) error {
	_, err := p.Write(data)
	return err
}

type TCPTransportOpts struct {
	ListenAddr string
}

// Implements the Transport interface
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC

	mu    sync.RWMutex // For Peers
	peers map[net.Addr]TCPPeer
}

func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Consume() <-chan RPC { // read-only channel of messages
	return t.rpcch
}

func (t *TCPTransport) Connect(addr string) {

}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go t.StartAcceptLoop()
	log.Printf("TCP transport listening on port: %s", t.ListenAddr)
	return nil
}

func (t *TCPTransport) StartAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			log.Printf("TCP accept error %s\n", err)
		}
		log.Printf("new incoming connection %+v\n", conn)

		go t.HandleConn(conn)
	}
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.HandleConn(conn)
	return nil
}

func (t *TCPTransport) HandleConn(conn net.Conn) {
	var err error
	defer func() {
		fmt.Printf("dropping peer connection: %s\n", err)
		conn.Close()
	}()
}
