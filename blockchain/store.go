package blockchain

type Store interface {
	// Block operations
	// Don't need a delete/update block method cause
	// a blockchain is immutable by design
	PutBlock(block *Block) error
	GetBlock(hash BlockHash) (*Block, error)

	// UTXO operations
	// Create/Update
	// PutUTXO(txID TransactionID, utxo *TxOutputs) error
	// Read
	GetUTXO(txID TransactionID) (*TxOutputs, error)
	// GetAllUTXOs() (UTXOs, error)
	// Delete
	DeleteUTXO(txID TransactionID) error
	DeleteAllUTXOs() error

	// Batch UTXO operations
	PutUTXOs(UTXOs) error
	// DeleteUTXOs() error

	Close() error
}
