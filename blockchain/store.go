package blockchain

import "github.com/dgraph-io/badger/v4"

// TODO: Implement persistent storage

type Store interface {
}

type NativeStore struct {
}

type BadgerStore struct {
	Encoder GobEncoderDecoder
	db      *badger.DB
}
