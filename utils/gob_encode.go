package utils

import (
	"bytes"
	"encoding/gob"
)

func EncodeInGob[T any](obj *T) ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(*obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeInGob[T any](data []byte) (*T, error) {
	var obj T
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}
