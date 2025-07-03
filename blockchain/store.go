package blockchain

type Store interface {
	// Block operations
	// Don't need a delete/update block method cause
	// a blockchain is immutable by design
	PutBlock(block *Block) error
	GetBlock(hash BlockHash) (*Block, error)

	// UTXO operations
	PutUTXO(txID TransactionID, utxo *TxOutputs) error
	GetUTXO(txID TransactionID) (*TxOutputs, error)
	GetAllUTXOs() (UTXOs, error)
	DeleteUTXO(txID TransactionID) error

	// Batch UTXO operations
	// PutUTXOs(UTXOs) error
	// DeleteUTXOs(UTXOs) error

	Close() error
}
