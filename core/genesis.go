package core

import (
	"fmt"
	"os"

	"github.com/pragmaticcs/blockchain/utils"
)

type GenesisConfig struct {
	Amount     int
	PubKeyHash []byte
}

func (g *GenesisConfig) ToFile(path string) error {
	buf, err := utils.EncodeInGob(g)
	if err != nil {
		return err
	}
	filePath := fmt.Sprintf("%s/genesis.data", path)
	return os.WriteFile(filePath, buf, 0644)
}

func LoadGenesisConfig(path string) (*GenesisConfig, error) {
	filePath := fmt.Sprintf("%s/genesis.data", path)
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	conf, err := utils.DecodeInGob[GenesisConfig](buf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
