package p2p

import "github.com/pragmaticcs/blockchain/core"

type Node struct {
	State  *core.StateManager
	Server *Server
}
