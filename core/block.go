package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"log"
	"time"

	"github.com/pragmaticcs/blockchain/crypto"
)

const CurrentVersion uint32 = 1

type BlockHash = []byte

type BlockHeader struct {
	Version    uint32
	PrevHash   BlockHash
	MerkleRoot MerkleHash
	Time       int64
	Nonce      int32
	Difficulty uint32
}

type BlockBody struct {
	Transactions []*Transaction
}

type Block struct {
	BlockHeader
	BlockBody
}

func NewBlock(prevHash []byte, transactions []*Transaction) (*Block, error) {
	blockBody := BlockBody{
		Transactions: transactions,
	}
	merkleTree, err := blockBody.GetMerkleTree()
	if err != nil {
		return nil, err
	}
	block := &Block{
		BlockHeader: BlockHeader{
			PrevHash:   prevHash,
			MerkleRoot: merkleTree.Root,
			Time:       time.Now().Unix(),
			Nonce:      0,
			Difficulty: 0,
		},
		BlockBody: blockBody,
	}
	return block, nil
}

func (b *BlockBody) GetMerkleTree() (*MerkleTree, error) {
	var txIDs []TransactionID
	for _, tx := range b.Transactions {
		txIDs = append(txIDs, tx.ID)
	}
	return NewMerkleTree(txIDs)
}

func (b *BlockHeader) GetData() []byte {
	// Converts block into bytes
	data := bytes.Join([][]byte{
		GetHexValue(int64(b.Version)),
		b.PrevHash[:],
		b.MerkleRoot,
		GetHexValue(b.Time),
		GetHexValue(int64(b.Nonce)),
		GetHexValue(int64(b.Difficulty))},
		[]byte{})
	return data
}

func (b *BlockHeader) GetHash() BlockHash {
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
	return NewBlock([]byte{}, []*Transaction{coinbase})
}
