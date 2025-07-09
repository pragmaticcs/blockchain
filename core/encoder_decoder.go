package core

type BlockEncoderDecoder interface {
	Encode(*Block) ([]byte, error)
	Decode([]byte) (*Block, error)
}

type UtxoEncoderDecoder interface {
	Encode(*TxOutput) ([]byte, error)
	Decode([]byte) (*TxOutput, error)
}
