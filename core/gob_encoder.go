package core

import (
	"bytes"
	"encoding/gob"
)

type GobEncoder[T any] struct {
}

func (enc GobEncoder[T]) Encode(obj *T) ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(*obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (enc GobEncoder[T]) Decode(data []byte) (*T, error) {
	var obj T
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}
