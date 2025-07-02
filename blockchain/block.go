package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"log"

	"github.com/pragmaticcs/blockchain/crypto"
)

type BlockHash = []byte

type BlockHeader struct {
	Hash     BlockHash // 32 bytes
	PrevHash BlockHash // 32 bytes
	// MerkelRoot [32]byte
	// Time       time.Time // ???
	Nonce int32
}

type BlockBody struct {
	Transactions []*Transaction
}

type Block struct {
	// BlockSize int32
	BlockHeader
	BlockBody
}

func NewBlock(prevHash []byte, transactions []*Transaction) *Block {
	block := &Block{
		BlockHeader: BlockHeader{
			PrevHash: prevHash,
			Nonce:    0, // ??? Maybe should be an arg? IDK
		},
		BlockBody: BlockBody{
			Transactions: transactions,
		},
	}
	block.Hash = block.GetHash()
	return block
}

func (b *Block) GetTransactionsData() []byte {
	// Returns the hash of an array of IDs in the block
	var ids [][]byte

	for _, tx := range b.Transactions {
		ids = append(ids, tx.ID[:])
	}
	flattenedIds := bytes.Join(ids, []byte{})
	hash := sha256.Sum256(flattenedIds)
	return hash[:]
}

func (b *Block) GetData() []byte {
	// Converts block into bytes
	data := bytes.Join([][]byte{
		b.PrevHash[:],
		b.GetTransactionsData()},
		[]byte{})
	return data
}

func (b *Block) GetHash() []byte {
	// Returns the hash of the block
	hash := sha256.Sum256(b.GetData())
	return hash[:]
}

func GetHexValue(i int64) []byte {
	// Converting an integer into binary
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, i)
	if err != nil {
		log.Panic(err) // Something has gone horribly wrong
	}
	return buf.Bytes()
}

func CreateGenesisBlock(value int, to crypto.PubKey) (*Block, error) {
	coinbase, err := NewCoinbase(int64(value), to)
	if err != nil {
		return nil, err
	}
	return NewBlock([]byte{}, []*Transaction{coinbase}), nil
}
