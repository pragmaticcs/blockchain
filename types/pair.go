package types

type Pair[K any, V any] struct {
	Key   K
	Value V
}

func NewPair[K any, V any](k K, v V) *Pair[K, V] {
	return &Pair[K, V]{
		Key:   k,
		Value: v,
	}
}
