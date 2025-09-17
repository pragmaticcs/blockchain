package core

type Store interface {
	// Block operations
	// Don't need a delete block method cause
	// a blockchain is immutable by design
	PutBlock(block *Block) error
	GetBlock(hash BlockHash) (*Block, error)
	BlockExists(hash BlockHash) bool
	TxExists(id TransactionID) bool
	GetBlockchainHead() (*Block, error)
	SetBlockchainHead(block *Block) error

	// UTXO operations
	ForEachUTXO(fn func(outPoint OutPoint, out TxOutput) error) error
	DeleteAllUTXOs() error

	PutUTXOs(UTXOs) error
	DeleteUTXOs([]OutPoint) error
	GetUTXO(OutPoint) (*TxOutput, error)
	// Combined operation for atomicity
	PutDeleteUTXOs(put UTXOs, delete []OutPoint) error

	Close() error
}
