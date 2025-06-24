package blockchain

import (
	"bytes"
	"encoding/gob"
)

type EncoderDecoder interface {
	EncodeBlock(*Block) []byte
	DecodeBlock([]byte) *Block
}

type GobEncoderDecoder struct{}

func (g *GobEncoderDecoder) EncodeBlock(b *Block) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(*b)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *GobEncoderDecoder) DecodeBlock(b []byte) (*Block, error) {
	var block Block
	err := gob.NewDecoder(bytes.NewBuffer(b)).Decode(&block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}
