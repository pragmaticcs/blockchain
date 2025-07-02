package blockchain

type BlockEncoder interface {
	Encode(*Block) ([]byte, error)
	Decode([]byte) (*Block, error)
}

type UtxoEncoder interface {
	Encode(*TxOutputs) ([]byte, error)
	Decode([]byte) (*TxOutputs, error)
}
