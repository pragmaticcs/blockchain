package p2p

import (
	"encoding/binary"
	"fmt"
	"io"
)

const messageMagic uint32 = 0xf4e7a92c

type RpcCommand uint16

const (
	RpcCommandVersion RpcCommand = iota
	RpcCommandAddr
	RpcCommandBlock
	RpcCommandTx
)

type MessageHeader struct {
	Magic   uint32
	Command RpcCommand
	Length  uint32
	// TODO: add checksum
	// Checksum [4]byte
}

type Message struct {
	Header  MessageHeader
	Payload []byte
}

func NewMessage(payload []byte, command RpcCommand) *Message {
	header := MessageHeader{
		Magic:   messageMagic,
		Command: command,
		Length:  uint32(len(payload)),
	}
	return &Message{
		Header:  header,
		Payload: payload,
	}
}

type MessageDecoder struct {
	r io.Reader
}

func NewMessageDecoder(r io.Reader) *MessageDecoder {
	return &MessageDecoder{
		r: r,
	}
}

func (d *MessageDecoder) Decode() (*Message, error) {
	var header MessageHeader

	err := binary.Read(d.r, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}

	magic := header.Magic
	if magic != messageMagic {
		return nil, NewMessageError(fmt.Sprintf("invalid magic number %x", header.Magic))
	}

	// TODO: Add a check for maximum payload length
	// to prevent allocating too much memory
	payload := make([]byte, header.Length)
	err = binary.Read(d.r, binary.LittleEndian, payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		Header:  header,
		Payload: payload,
	}, nil
}

type MessageEncoder struct {
	w io.Writer
}

func NewMessageEncoder(w io.Writer) *MessageEncoder {
	return &MessageEncoder{
		w: w,
	}
}

func (e *MessageEncoder) Encode(msg *Message) error {
	err := binary.Write(e.w, binary.LittleEndian, msg.Header)
	if err != nil {
		return err
	}
	_, err = e.w.Write(msg.Payload)
	return err
}

type MessageError struct {
	Err string
}

func NewMessageError(e string) MessageError {
	return MessageError{Err: e}
}

func (e MessageError) Error() string {
	return fmt.Sprintf("Message error: %s", e.Err)
}
