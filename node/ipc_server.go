package node

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

type IpcCommand string

const (
	IpcGetBalance IpcCommand = "get_balance"
	IpcStopNode   IpcCommand = "stop_node"
)

type IpcRequest struct {
	Command IpcCommand      `json:"command"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type IpcResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type IpcServer struct {
	node       *Node
	listener   net.Listener
	socketPath string
}

func NewIpcServer(path string, node *Node) (*IpcServer, error) {
	socketPath := fmt.Sprintf("%s/gochain.sock", path)

	err := os.RemoveAll(socketPath)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	log.Printf("ipc server: listening on %s", socketPath)

	return &IpcServer{
		node:       node,
		listener:   listener,
		socketPath: socketPath,
	}, nil
}

func (s *IpcServer) Start() {
	for {
		conn, err := s.listener.Accept()

		if err != nil {
			log.Panicf("ipc server: error during accept: %s", err)
		}
		go s.HandleConn(conn)
	}
}

func (s *IpcServer) HandleConn(conn net.Conn) {
	defer conn.Close()

	var req IpcRequest
	var resp *IpcResponse

	err := json.NewDecoder(conn).Decode(&req)
	if err != nil {
		log.Printf("ipc server: failed to decode request: %s", err)
		resp = &IpcResponse{
			Success: false,
			Error:   "invalid request",
		}
	} else {
		resp = s.handleRequest(&req)
	}

	err = json.NewEncoder(conn).Encode(resp)
	if err != nil {
		log.Printf("ipc server: failed to write response: %s", err)
	}
}

func (s *IpcServer) handleRequest(req *IpcRequest) *IpcResponse {
	switch req.Command {
	case IpcGetBalance:
		balance, _ := s.node.StateManager.GetBalance()
		return &IpcResponse{
			Success: true,
			Data:    balance,
		}
	case IpcStopNode:
		go s.node.sendQuitSignal()
		s.node.Stop()
		return &IpcResponse{
			Success: true,
			Data:    "node shutting down.",
		}
	default:
		return &IpcResponse{
			Success: false,
			Error:   fmt.Sprintf("unknown command: %s", req.Command),
		}
	}
}

func (s *IpcServer) Stop() {
	s.listener.Close()
}
