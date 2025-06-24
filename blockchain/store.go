package blockchain

// TODO: Implement persistent storage

type Store interface {
}

type NativeStore struct {
}

type BadgerStore struct {
	Encoder GobEncoderDecoder
	// db      *badger.DB
}
