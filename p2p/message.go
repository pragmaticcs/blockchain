package p2p

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pragmaticcs/blockchain/core"
	"github.com/pragmaticcs/blockchain/utils"
)

const (
	msgMagic            uint32 = 0xf4e7a92c
	msgPayloadMaxLength uint32 = 5 * 1024 * 1024
)

type MessageHeader struct {
	Magic   uint32
	Command RpcCommand
	Length  uint32
	// TODO: add checksum for the payload
	// Checksum [4]byte
}

type Message struct {
	// Message is what is sent through the network
	Header  MessageHeader
	Payload []byte
}

func NewMessage(payload []byte, command RpcCommand) *Message {
	header := MessageHeader{
		Magic:   msgMagic,
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
	if magic != msgMagic {
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

// Messages:
type MessageVersion struct {
	ProtocolVersion   uint32
	LatestBlockHeight uint32
}

func (msg *MessageVersion) ToBytes() ([]byte, error) {
	return utils.EncodeInGob(msg)
}

func (rpc *RPC) VisitAsMessageVersion() (*MessageVersion, error) {
	return utils.DecodeInGob[MessageVersion](rpc.Payload)
}

func (msg *Message) VisitAsMessageVersion() (*MessageVersion, error) {
	return utils.DecodeInGob[MessageVersion](msg.Payload)
}

type MessageAddr struct {
	AddrList []string
}

func (msg *MessageAddr) ToBytes() ([]byte, error) {
	return utils.EncodeInGob(msg)
}

func (rpc *RPC) VisitAsMessageAddr() (*MessageAddr, error) {
	return utils.DecodeInGob[MessageAddr](rpc.Payload)
}

func (msg *Message) VisitAsMessageAddr() (*MessageAddr, error) {
	return utils.DecodeInGob[MessageAddr](msg.Payload)
}

type MessageGetBlocks struct {
	From core.BlockHash
}

func (msg *MessageGetBlocks) ToBytes() ([]byte, error) {
	return utils.EncodeInGob(msg)
}

func (rpc *RPC) VisitAsMessageGetBlocks() (*MessageGetBlocks, error) {
	return utils.DecodeInGob[MessageGetBlocks](rpc.Payload)
}

func (msg *Message) VisitAsMessageGetBlocks() (*MessageGetBlocks, error) {
	return utils.DecodeInGob[MessageGetBlocks](msg.Payload)
}

type InvType uint8

const (
	InvTypeTx InvType = iota
	InvTypeBlock
)

type InvVector struct {
	Type InvType
	Hash []byte
}

type MessageInv struct {
	Items []InvVector
}

func (msg *MessageInv) ToBytes() ([]byte, error) {
	return utils.EncodeInGob(msg)
}

func (rpc *RPC) VisitAsMessageInv() (*MessageInv, error) {
	return utils.DecodeInGob[MessageInv](rpc.Payload)
}
func (msg *Message) VisitAsMessageInv() (*MessageInv, error) {
	return utils.DecodeInGob[MessageInv](msg.Payload)
}

// Requesting data is the same as announcing you have that data
type MessageGetData = MessageInv

type MessageBlock struct {
	Block core.Block
}

func (msg *MessageBlock) ToBytes() ([]byte, error) {
	return utils.EncodeInGob(msg)
}

func (rpc *RPC) VisitAsMessageBlock() (*MessageBlock, error) {
	return utils.DecodeInGob[MessageBlock](rpc.Payload)
}

func (msg *Message) VisitAsMessageBlock() (*MessageBlock, error) {
	return utils.DecodeInGob[MessageBlock](msg.Payload)
}

type MessageTx struct {
	Tx core.Transaction
}

func (msg *MessageTx) ToBytes() ([]byte, error) {
	return utils.EncodeInGob(msg)
}

func (rpc *RPC) VisitAsMessageTx() (*MessageTx, error) {
	return utils.DecodeInGob[MessageTx](rpc.Payload)
}

func (msg *Message) VisitAsMessageTx() (*MessageTx, error) {
	return utils.DecodeInGob[MessageTx](msg.Payload)
}

// Errors for the message module
type MessageError struct {
	Err string
}

func NewMessageError(e string) MessageError {
	return MessageError{Err: e}
}

func (e MessageError) Error() string {
	return fmt.Sprintf("Message error: %s", e.Err)
}
