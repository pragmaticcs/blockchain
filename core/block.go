package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"log"
	"time"
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
	Height     uint32
}

type BlockBody struct {
	Transactions []Transaction
}

type Block struct {
	BlockHeader
	BlockBody
}

func NewBlock(prevHash []byte, transactions []Transaction, prevHeight uint32) (*Block, error) {
	blockBody := BlockBody{
		Transactions: transactions,
	}
	merkleTree, err := blockBody.GetMerkleTree()
	if err != nil {
		return nil, err
	}
	block := &Block{
		BlockHeader: BlockHeader{
			Version:    CurrentVersion,
			PrevHash:   prevHash,
			MerkleRoot: merkleTree.Root,
			Time:       time.Now().Unix(),
			Nonce:      0,
			Difficulty: 0,
			Height:     prevHeight + 1,
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
	// Converts block header into bytes
	data := bytes.Join([][]byte{
		GetHexValue(int64(b.Version)),
		b.PrevHash[:],
		b.MerkleRoot,
		GetHexValue(b.Time),
		GetHexValue(int64(b.Nonce)),
		GetHexValue(int64(b.Difficulty)),
		GetHexValue(int64(b.Height))},
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

func CreateGenesisBlock(conf GenesisConfig) (*Block, error) {
	coinbase := NewCoinbase(conf.Amount, conf.PubKeyHash)
	blockBody := BlockBody{
		Transactions: []Transaction{*coinbase},
	}
	merkleTree, err := blockBody.GetMerkleTree()
	if err != nil {
		return nil, err
	}
	return &Block{
		BlockHeader: BlockHeader{
			Version:    CurrentVersion,
			PrevHash:   []byte{},
			MerkleRoot: merkleTree.Root,
			Time:       conf.Time,
			Nonce:      0,
			Difficulty: 0,
			Height:     0,
		},
		BlockBody: blockBody,
	}, nil
}
