package core

import (
	"fmt"
	"sync"

	"github.com/pragmaticcs/blockchain/crypto"
	"github.com/pragmaticcs/blockchain/wallet"
)

// Just for testing purposes
const (
	DefaultDifficulty = 10
)

type StateManager struct {
	Blockchain *Blockchain
	UtxoSet    *UTXOset
	Wallet     *wallet.Wallet
	// Copy of pointer to store to handle closing
	store Store

	miningDifficulty uint32

	// Mutexes
	lock sync.RWMutex // Use when modifying the state of blockchain itself
}

func NewStateManager(b *Blockchain, w *wallet.Wallet) *StateManager {
	utxoSet := NewUTXOset(b, b.Store)
	utxoSet.Reindex()
	return &StateManager{
		Blockchain:       b,
		Wallet:           w,
		UtxoSet:          utxoSet,
		store:            b.Store,
		miningDifficulty: DefaultDifficulty,
		lock:             sync.RWMutex{},
	}
}

func (s *StateManager) Close() error {
	return s.store.Close()
}

func (s *StateManager) GetLatestBlockHeight() uint32 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.Blockchain.CurrHeight
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

func (s *StateManager) NewTransaction(to wallet.Address, amount int, fee int) (*Transaction, error) {
	var inputs []TxInput
	var outputs []TxOutput

	selfAddress, err := s.Wallet.GetAddress()
	if err != nil {
		return nil, err
	}
	pubKeyHash, err := crypto.HashPublicKey(s.Wallet.PublicKey)
	if err != nil {
		return nil, err
	}
	sum, outPoints := s.UtxoSet.FindSpendableOutputs(pubKeyHash, amount+fee)
	if sum < int64(amount) {
		return nil, wallet.ErrorInsufficientBalance
	}
	for _, outPoint := range outPoints {
		in := TxInput{
			TxID:   outPoint.TxID,
			OutID:  outPoint.OutID,
			Sig:    nil,
			PubKey: s.Wallet.PublicKey,
		}
		inputs = append(inputs, in)
	}
	outputs = append(outputs, *NewTxOutput(amount, []byte(to))) // Transfer to reciever
	if sum > int64(amount) {
		outputs = append(outputs, *NewTxOutput(int(sum)-amount, []byte(selfAddress))) // Transfer change back
	}

	tx := &Transaction{
		ID:      TransactionZeroID,
		Outputs: outputs,
		Inputs:  inputs,
	}
	txID := tx.GetHash()
	tx.ID = txID

	err = s.Blockchain.SignTransaction(tx, s.Wallet.PrivateKey)

	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *StateManager) Mine(transactions []Transaction) (*Block, error) {
	block, err := s.Blockchain.NewBlock(transactions)
	if err != nil {
		return nil, err
	}
	// Run the proof of work
	// TODO: Implement a calculateDifficulty function
	pow := NewPOW(s.miningDifficulty)

	// TODO: handle if couldn't mine the block
	_ = pow.Run(block)

	return block, err
}

func (s *StateManager) BlockExists(hash BlockHash) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.store.Exists(hash)
}

func (s *StateManager) GetBlock(hash BlockHash) (*Block, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.store.GetBlock(hash)
}

func (s *StateManager) updateUnlocked(block *Block) error {
	err := s.Blockchain.AddBlock(block)
	if err != nil {
		return err
	}
	return s.UtxoSet.Update(block)
}

func (s *StateManager) Update(block *Block) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.updateUnlocked(block)
}

func (s *StateManager) Send(to wallet.Address, amount int, fee int) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	tx, err := s.NewTransaction(to, amount, fee)
	if err != nil {
		return err
	}
	fmt.Println("Mining new block with transactions")
	block, err := s.Mine([]Transaction{*tx})
	if err != nil {
		return err
	}
	return s.updateUnlocked(block)
}

func (s *StateManager) CalculateFee(tx *Transaction) (int, error) {
	inputs := 0
	for _, in := range tx.Inputs {
		outPoint := NewOutPoint(in.TxID, in.OutID)
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
