package node

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type IpcCommand uint16

const (
	IpcCommandGetBalance IpcCommand = iota
	IpcCommandStop
	IpcCommandSend
)

type IPC struct {
	Command IpcCommand
	Payload []byte
}

type IpcServer struct {
	listener   net.Listener
	socketPath string
	ipcCh      chan IPC
}

func NewIpcServer(path string) (*IpcServer, error) {
	socketPath := fmt.Sprintf("%s/mychain.sock", path)

	err := os.RemoveAll(socketPath)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	log.Printf("new ipc server listening on %s", socketPath)

	return &IpcServer{
		listener:   listener,
		socketPath: socketPath,
		ipcCh:      make(chan IPC),
	}, nil
}

func (s *IpcServer) Start() {
	for {
		conn, err := s.listener.Accept()

		if err != nil {
			log.Panicf("ipc server error during accept: %s", err)
		}
		go s.HandleConn(conn)
	}
}

func (s *IpcServer) HandleConn(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// TODO: Handle the command from the conn
}

func (s *IpcServer) Stop() {
	s.listener.Close()
}

func (s *IpcServer) Consume() <-chan IPC {
	return s.ipcCh
}
