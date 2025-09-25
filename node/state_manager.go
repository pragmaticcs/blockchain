package node

import (
	"fmt"
	"sync"

	"github.com/pragmaticcs/blockchain/core"
	"github.com/pragmaticcs/blockchain/crypto"
	"github.com/pragmaticcs/blockchain/wallet"
)

// Just for testing purposes
const (
	DefaultDifficulty = 28
)

type StateManager struct {
	Blockchain *core.Blockchain
	UtxoSet    *core.UTXOset
	Wallet     *wallet.Wallet
	Mempool    *Mempool

	// Mutexes
	lock sync.RWMutex // Use when modifying the state of blockchain itself
}

func NewStateManager(b *core.Blockchain, w *wallet.Wallet, m *Mempool) (*StateManager, error) {
	utxoSet := core.NewUTXOset(b, b.Store)
	err := utxoSet.Reindex()
	if err != nil {
		return nil, err
	}
	return &StateManager{
		Blockchain: b,
		Wallet:     w,
		UtxoSet:    utxoSet,
		Mempool:    m,
		lock:       sync.RWMutex{},
	}, nil
}

func (s *StateManager) Close() error {
	return s.Blockchain.Close()
}

func (s *StateManager) GetLatestBlockHeight() uint32 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.Blockchain.CurrHeight
}

func (s *StateManager) GetBlockchainHead() core.BlockHash {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.Blockchain.Head
}

func (s *StateManager) GetBalance() (int, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	pubKeyHash, err := crypto.HashPublicKey(s.Wallet.PublicKey)
	if err != nil {
		return -1, err
	}
	return s.UtxoSet.FindSpendableAmount(pubKeyHash)
}

func (s *StateManager) NewTransaction(to wallet.Address, amount, fee int) (*core.Transaction, error) {
	var inputs []core.TxInput
	var outputs []core.TxOutput

	reservedOuts, availableOuts := s.Mempool.GetReservedOutputs(s.Wallet.PublicKey)
	unconfirmedTxs := s.Mempool.GetTransactionTable()

	selfAddress, err := s.Wallet.GetAddress()
	if err != nil {
		return nil, err
	}
	pubKeyHash, err := crypto.HashPublicKey(s.Wallet.PublicKey)
	if err != nil {
		return nil, err
	}
	needed := int64(amount + fee)

	sum, outPoints := s.UtxoSet.FindSpendableOutputs(pubKeyHash, int(needed), reservedOuts, availableOuts)
	if sum < needed {
		return nil, wallet.ErrorInsufficientBalance
	}
	for _, outPoint := range outPoints {
		in := core.TxInput{
			TxID:   outPoint.TxID,
			OutID:  outPoint.OutID,
			Sig:    nil,
			PubKey: s.Wallet.PublicKey,
		}
		inputs = append(inputs, in)
	}
	outputs = append(outputs, *core.NewTxOutput(int64(amount), []byte(to))) // Transfer to reciever

	if sum > needed {
		outputs = append(outputs, *core.NewTxOutput(sum-needed, []byte(selfAddress))) // Transfer change back
	}

	tx := &core.Transaction{
		ID:      core.TransactionZeroID,
		Outputs: outputs,
		Inputs:  inputs,
	}

	txID := tx.GetHash()
	tx.ID = txID

	err = s.Blockchain.SignTransaction(tx, s.Wallet.PrivateKey, unconfirmedTxs)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *StateManager) Update(block *core.Block) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	err := s.Blockchain.AddBlock(block)
	if err != nil {
		return err
	}
	return s.UtxoSet.Update(block)
}

func (s *StateManager) Validate(block *core.Block) (bool, error) {
	pow := core.NewPOW(block.Difficulty)
	valid := pow.Validate(block)
	if !valid {
		return false, nil
	}
	for _, tx := range block.Transactions {
		if tx.IsCoinbase() {
			// TODO: Check if mining reward is the default one
			continue
		}
		valid, err := s.Blockchain.VerifyTransaction(&tx, nil)
		if err != nil {
			return false, err
		}
		if !valid {
			return false, nil
		}
	}
	return true, nil
}

func (s *StateManager) CalculateFee(tx *core.Transaction) (int, error) {
	inputs := 0
	for _, in := range tx.Inputs {
		outPoint := core.NewOutPoint(in.TxID, in.OutID)

		out, err := s.UtxoSet.Store.GetUTXO(outPoint)
		if err != nil {
			return -1, err
		}
		if out == nil {
			return -1, fmt.Errorf("output is either spent or doesn't exist")
		}
		inputs += int(out.Value)
	}

	outputs := 0
	for _, out := range tx.Outputs {
		outputs += int(out.Value)
	}
	if inputs < outputs {
		return -1, fmt.Errorf("sum of inputs can't be less than sum of outputs")
	}
	return inputs - outputs, nil
}
