package blockchain

type Store interface {
	// Block operations
	// Don't need a delete/update block method cause
	// a blockchain is immutable by design
	PutBlock(block *Block) error
	GetBlock(hash BlockHash) (*Block, error)

	// UTXO operations
	ForEachUTXO(fn func(outPoint OutPoint, out TxOutput) error) error
	DeleteAllUTXOs() error

	PutUTXOs(UTXOs) error
	DeleteUTXOs([]OutPoint) error

	Close() error
}
