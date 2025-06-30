package blockchain

type BlockEncoder interface {
	Encode(*Block) ([]byte, error)
	Decode([]byte) (*Block, error)
}
