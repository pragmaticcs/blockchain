package blockchain

import (
	"sync"

	"github.com/pragmaticcs/blockchain/crypto"
	"github.com/pragmaticcs/blockchain/wallet"
)

type StateManager struct {
	Blockchain *BlockChain
	Wallets    *wallet.Wallets

	// Mutexes
	selfLock    sync.RWMutex // Use when modifying the state of blockchain itself
	accountLock sync.RWMutex // Use when modifying the state of accounts/wallets
}

func NewStateManager(b *BlockChain, w *wallet.Wallets) *StateManager {
	return &StateManager{
		Blockchain:  b,
		Wallets:     w,
		selfLock:    sync.RWMutex{},
		accountLock: sync.RWMutex{},
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

func (s *StateManager) GetBalance(addr wallet.Address) (int, error) {
	wallet, err := s.Wallets.GetWallet(addr)
	if err != nil {
		return -1, err
	}
	pubKeyHash, err := crypto.HashPublicKey(wallet.PublicKey)
	if err != nil {
		return -1, err
	}
	return s.Blockchain.FindSpendableAmount(pubKeyHash)

}
