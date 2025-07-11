package core

type Blockchain struct {
	Store      Store
	Head       BlockHash
	CurrHeight uint32
}

func InitBlockchain(store Store, genesisConfDir string) (*Blockchain, error) {
	// TODO: Handle fetching the blockchain from the store
	// and setting the CurrHeight
	head, err := store.GetBlockchainHead()
	if err != nil {
		return nil, err
	}
	if head == nil {
		genesis, err := LoadGenesisConfig(genesisConfDir)
		if err != nil {
			return nil, err
		}
		return NewBlockchain(*genesis, store)
	}
	return &Blockchain{
		Store:      store,
		Head:       head.GetHash(),
		CurrHeight: head.Height,
	}, nil
}

func NewBlockchain(genesisConfig GenesisConfig, store Store) (*Blockchain, error) {
	genesis, err := CreateGenesisBlock(genesisConfig.Amount, genesisConfig.PubKeyHash)
	if err != nil {
		return nil, err
	}
	store.PutBlock(genesis)
	store.SetBlockchainHead(genesis)
	return &Blockchain{
		Store:      store,
		Head:       genesis.GetHash(),
		CurrHeight: 1, // Because of the genesis block
	}, nil
}

func (b *Blockchain) Iterate() func(func(*Block) bool) {
	return func(yield func(*Block) bool) {
		currHash := b.Head
		for currHash != nil {
			currBlock, err := b.Store.GetBlock(currHash)
			if err != nil {
				return
			}
			if !yield(currBlock) {
				return
			}
			currHash = currBlock.PrevHash
		}
	}
}

func (b *Blockchain) NewBlock(transactions []Transaction) (*Block, error) {
	return NewBlock(b.Head, transactions, b.CurrHeight)
}

func (b *Blockchain) AddBlock(block *Block) error {
	b.Head = block.GetHash()
	b.CurrHeight++
	err := b.Store.PutBlock(block)
	if err != nil {
		return err
	}
	return b.Store.SetBlockchainHead(block)
}
