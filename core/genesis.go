package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type GenesisConfig struct {
	Amount     int64
	PubKeyHash []byte
}

func (g *GenesisConfig) ToFile(path string) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, g.Amount)
	if err != nil {
		return err
	}

	_, err = buf.Write(g.PubKeyHash)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/genesis.data", path)
	return os.WriteFile(filePath, buf.Bytes(), 0644)
}

func LoadGenesisConfig(path string) (*GenesisConfig, error) {
	filePath := fmt.Sprintf("%s/genesis.data", path)
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var genesis GenesisConfig
	reader := bytes.NewReader(buf)

	err = binary.Read(reader, binary.BigEndian, &genesis.Amount)
	if err != nil {
		return nil, err
	}

	genesis.PubKeyHash = make([]byte, reader.Len())
	_, err = reader.Read(genesis.PubKeyHash)
	if err != nil {
		return nil, err
	}
	return &genesis, nil
}
