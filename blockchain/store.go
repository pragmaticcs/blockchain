package blockchain

type Store interface {
	PutBlock(block *Block) error
	GetBlock(hash BlockHash) (*Block, error)

	// // UTXO operations
	// PutUTXO(txID TransactionID, outs []TxOutput) error
	// GetUTXO(txID TransactionID) ([]TxOutput, error)
	// DeleteUTXO(txID TransactionID) error

	Close() error
}
