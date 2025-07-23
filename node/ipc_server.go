package node

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/pragmaticcs/blockchain/wallet"
)

type IpcCommand string

const (
	IpcGetBalance IpcCommand = "get_balance"
	IpcStopNode   IpcCommand = "stop_node"
	IpcSend       IpcCommand = "send"
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
	quitCh     chan struct{}
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
		quitCh:     make(chan struct{}),
	}, nil
}

func (s *IpcServer) Start() {
	for {
		select {
		case <-s.quitCh:
			s.Stop()
			return
		default:
			conn, err := s.listener.Accept()

			if err != nil {
				log.Panicf("ipc server: error during accept: %s", err)
			}
			go s.HandleConn(conn)
		}
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
		defer s.node.sendQuitSignal()
		return &IpcResponse{
			Success: true,
			Data:    "node shutting down.",
		}
	case IpcSend:
		var payload IpcPayloadSend
		err := json.Unmarshal(req.Payload, &payload)
		if err != nil {
			return &IpcResponse{
				Success: false,
				Error:   "couldn't parse payload",
			}
		}
		err = s.node.Send(wallet.Address(payload.Address), int(payload.Amount), int(payload.Fee))
		if err != nil {
			return &IpcResponse{
				Success: false,
				Error:   fmt.Sprintf("couln't execute transaction: %s", err),
			}
		}
		return &IpcResponse{
			Success: true,
			Data:    "successful transaction broadcast",
		}
	default:
		return &IpcResponse{
			Success: false,
			Error:   fmt.Sprintf("unknown command: %s", req.Command),
		}
	}
}

func (s *IpcServer) sendQuitCh() {
	s.quitCh <- struct{}{}
}

func (s *IpcServer) Stop() {
	s.listener.Close()
}

type IpcPayloadSend struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
	Fee     uint64 `json:"fee"`
}
