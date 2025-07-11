package utils

import (
	"encoding/binary"
)

func EncodeInBinary[T any](obj *T) []byte {
	buf := make([]byte, binary.Size(*obj))
	_, _ = binary.Encode(buf, binary.LittleEndian, *obj)
	return buf
}

func DecodeInBinary[T any](data []byte) (*T, error) {
	var obj T
	_, err := binary.Decode(data, binary.LittleEndian, &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}
