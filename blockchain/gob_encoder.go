package blockchain

import (
	"bytes"
	"encoding/gob"
)

type GobBlockEncoder struct {
}

func (enc *GobBlockEncoder) Encode(block *Block) ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(*block)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (enc *GobBlockEncoder) Decode(data []byte) (*Block, error) {
	var block Block
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}
