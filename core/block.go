package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"log"

	"github.com/pragmaticcs/blockchain/crypto"
)

type BlockHash = []byte

type BlockHeader struct {
	Hash       BlockHash // 32 bytes
	PrevHash   BlockHash // 32 bytes
	MerkelRoot MerkleHash
	// Time       time.Time // ???
	Nonce int32
}

type BlockBody struct {
	Transactions []*Transaction
}

type Block struct {
	BlockHeader
	BlockBody
}

func NewBlock(prevHash []byte, transactions []*Transaction) *Block {
	blockBody := BlockBody{
		Transactions: transactions,
	}
	merkleTree := blockBody.GetMerkleTree()
	block := &Block{
		BlockHeader: BlockHeader{
			PrevHash:   prevHash,
			Nonce:      0,
			MerkelRoot: merkleTree.Root,
		},
		BlockBody: blockBody,
	}
	return block
}

func (b *BlockBody) GetMerkleTree() *MerkleTree {
	var txIDs []TransactionID
	for _, tx := range b.Transactions {
		txIDs = append(txIDs, tx.ID)
	}
	merkleTree, _ := NewMerkleTree(txIDs)
	return merkleTree
}

func (b *BlockHeader) GetData() []byte {
	// Converts block into bytes
	data := bytes.Join([][]byte{
		b.PrevHash[:],
		b.MerkelRoot,
		GetHexValue(int64(b.Nonce))},
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
	return NewBlock([]byte{}, []*Transaction{coinbase}), nil
}
