package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"log"
)

type BlockHeader struct {
	// Version    uint32
	Hash     [32]byte
	PrevHash [32]byte
	// MerkelRoot [32]byte
	// Time       time.Time // ???
	// Bits  int32
	Nonce int32
}

type Block struct {
	// BlockSize int32
	BlockHeader
	BlockBody []byte // Data of the Block (Transactions in the future)
}

func NewBlock(prevHash [32]byte, body []byte) *Block {
	block := &Block{
		BlockHeader: BlockHeader{
			PrevHash: prevHash,
			Nonce:    0, // ??? Maybe should be an arg? IDK
		},
		BlockBody: body,
	}
	block.Hash = block.GetHash()
	return block
}

func (b *Block) GetData() []byte {
	data := bytes.Join([][]byte{
		b.PrevHash[:],
		b.BlockBody},
		[]byte{})
	return data
}

func (b *Block) GetHash() [32]byte {
	return sha256.Sum256(b.GetData())
}

func GetHexValue(i int32) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, i)
	if err != nil {
		log.Panic(err) // Something has gone horribly wrong
	}
	return buf.Bytes()
}

func CreateGenesisBlock() *Block {
	return NewBlock([32]byte{}, []byte("genesis"))
}
