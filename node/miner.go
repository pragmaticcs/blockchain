package node

import (
	"context"
	"log"

	"github.com/pragmaticcs/blockchain/core"
	"github.com/pragmaticcs/blockchain/crypto"
	"github.com/pragmaticcs/blockchain/p2p"
)

type Miner struct {
	Node           *Node
	currCtx        context.Context
	currCancelFunc context.CancelFunc
	isMining       bool
	startMining    chan struct{}
}

func NewMiner(node *Node) *Miner {
	return &Miner{
		Node:        node,
		startMining: make(chan struct{}),
		isMining:    false,
	}
}

func (m *Miner) Start() {
	go m.MineLoop()
}

func (m *Miner) StartMining() {
	select {
	case m.startMining <- struct{}{}:
	default:
	}
}

func (m *Miner) Stop() {
	close(m.startMining)
}

func (m *Miner) MineLoop() {
	for {
		<-m.startMining

		if m.isMining {
			continue
		}

		if len(m.Node.Mempool.GetTransactions()) == 0 {
			continue
		}

		m.currCtx, m.currCancelFunc = context.WithCancel(context.Background())

		go func() {
			err := m.Mine(m.currCtx)
			if err != nil {
				log.Printf("error during mining: %s", err)
			}
		}()
	}
}

func (m *Miner) Reset() {
	if m.currCancelFunc != nil {
		log.Println("miner resetting mining process")
		m.currCancelFunc()
	}
}

func (m *Miner) Mine(ctx context.Context) error {
	log.Println("miner started mining new block")

	m.isMining = true
	defer func() {
		m.isMining = false
	}()

	mempoolTxs := m.Node.Mempool.GetTransactions()

	// Add one to the length for the coinbase mining reward
	transactions := make([]core.Transaction, 0, len(mempoolTxs)+1)

	// Calculate mining fees
	miningFees := 0
	for _, tx := range mempoolTxs {
		fee, err := m.Node.StateManager.CalculateFee(tx)
		if err != nil {
			return err
		}
		miningFees += fee
	}

	ownPubKeyHash, err := crypto.HashPublicKey(m.Node.StateManager.Wallet.PublicKey)
	if err != nil {
		return err
	}
	coinbase := core.NewCoinbase(int64(DefaultMineReward+miningFees), ownPubKeyHash)

	transactions = append(transactions, *coinbase)

	for _, tx := range mempoolTxs {
		transactions = append(transactions, *tx)
	}

	pow := core.NewPOW(DefaultDifficulty)
	block, err := m.Node.StateManager.Blockchain.NewBlock(transactions)
	if err != nil {
		return err
	}

	if !pow.Run(ctx, block) {
		// TODO: Handle this better
		return nil
	}
	log.Println("miner successfully mined block")

	err = m.Node.StateManager.Update(block)
	if err != nil {
		return err
	}

	m.Node.Mempool.Clear(block)
	m.Node.BroadcastInv(block.GetHash(), p2p.InvTypeBlock)

	return nil
}
