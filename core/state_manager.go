package core

import (
	"fmt"
	"sync"

	"github.com/pragmaticcs/blockchain/crypto"
	"github.com/pragmaticcs/blockchain/wallet"
)

// Just for testing purposes
const (
	defaultDifficulty = 20
)

type StateManager struct {
	Blockchain *Blockchain
	UtxoSet    *UTXOset
	Wallets    *wallet.Wallets
	// Copy of pointer to store to handle closing
	store Store

	miningDifficulty uint32

	// Mutexes
	selfLock    sync.RWMutex // Use when modifying the state of blockchain itself
	accountLock sync.RWMutex // Use when modifying the state of accounts/wallets
}

func NewStateManager(b *Blockchain, w *wallet.Wallets) *StateManager {
	utxoSet := NewUTXOset(b, b.Store)
	utxoSet.Reindex()
	return &StateManager{
		Blockchain:       b,
		Wallets:          w,
		UtxoSet:          utxoSet,
		store:            b.Store,
		miningDifficulty: defaultDifficulty,
		selfLock:         sync.RWMutex{},
		accountLock:      sync.RWMutex{},
	}
}

func (s *StateManager) Lock() {
	s.selfLock.Lock()
}

func (s *StateManager) Unlock() {
	s.selfLock.Unlock()
}

func (s *StateManager) LockAccounts() {
	s.accountLock.Lock()
}

func (s *StateManager) UnlockAccounts() {
	s.accountLock.Unlock()
}

func (s *StateManager) Close() error {
	return s.store.Close()
}

func (s *StateManager) GetLatestBlockHeight() uint32 {
	return s.Blockchain.CurrHeight
}

func (s *StateManager) GetBalance(addr wallet.Address) (int, error) {
	wallet, err := s.Wallets.GetWallet(addr)
	if err != nil {
		return -1, err
	}
	pubKeyHash, err := crypto.HashPublicKey(wallet.PublicKey)
	if err != nil {
		return -1, err
	}
	return s.UtxoSet.FindSpendableAmount(pubKeyHash)

}

func (s *StateManager) NewTransaction(from, to wallet.Address, amount int) (*Transaction, error) {
	var inputs []TxInput
	var outputs []TxOutput

	sender, err := s.Wallets.GetWallet(from)
	if err != nil {
		return nil, err
	}
	pubKeyHash, err := crypto.HashPublicKey(sender.PublicKey)
	if err != nil {
		return nil, err
	}
	sum, outPoints := s.UtxoSet.FindSpendableOutputs(pubKeyHash, amount)
	if sum < int64(amount) {
		return nil, wallet.ErrorInsufficientBalance
	}
	for _, outPoint := range outPoints {
		in := TxInput{
			TxID:   outPoint.TxID,
			OutID:  outPoint.OutID,
			Sig:    nil,
			PubKey: sender.PublicKey,
		}
		inputs = append(inputs, in)
	}
	outputs = append(outputs, *NewTxOutput(amount, []byte(to))) // Transfer to reciever
	if sum > int64(amount) {
		outputs = append(outputs, *NewTxOutput(int(sum)-amount, []byte(from))) // Transfer change back
	}

	tx := &Transaction{
		ID:      TransactionZeroID,
		Outputs: outputs,
		Inputs:  inputs,
	}
	txID := tx.GetHash()
	tx.ID = txID

	err = s.Blockchain.SignTransaction(tx, sender.PrivateKey)

	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *StateManager) Mine(transactions []*Transaction) (*Block, error) {
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

func (s *StateManager) Update(block *Block) error {
	err := s.Blockchain.AddBlock(block)
	if err != nil {
		return err
	}
	return s.UtxoSet.Update(block)
}

func (s *StateManager) Send(from, to wallet.Address, amount int) error {
	tx, err := s.NewTransaction(from, to, amount)
	if err != nil {
		return err
	}
	fmt.Println("Mining new block with transactions")
	block, err := s.Mine([]*Transaction{tx})
	if err != nil {
		return err
	}
	return s.Update(block)
}
