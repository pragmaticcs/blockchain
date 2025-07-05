package core

type BlockEncoder interface {
	Encode(*Block) ([]byte, error)
	Decode([]byte) (*Block, error)
}

type UtxoEncoder interface {
	Encode(*TxOutput) ([]byte, error)
	Decode([]byte) (*TxOutput, error)
}
