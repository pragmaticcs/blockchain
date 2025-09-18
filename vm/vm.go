package vm

import (
	"fmt"
	"math/big"
)

const (
	StackSize = 256
)

type Value [32]byte

var opcodePushLengths = map[OpCode]int{
	OP_PUSH1: 1,
	OP_PUSH4: 4,
	OP_PUSH8: 8,
}

type VM struct {
	Chunk *Chunk
	Ip    int
	Stack *Stack
}

func NewVM(c *Chunk) *VM {
	return &VM{
		Chunk: c,
		Ip:    0,
		Stack: NewStack(),
	}
}

type Stack struct {
	Stack    []Value
	StackTop int
}

func NewStack() *Stack {
	return &Stack{
		Stack:    make([]Value, StackSize),
		StackTop: 0,
	}
}

func (s *Stack) Pop() Value {
	s.StackTop--
	return s.Stack[s.StackTop]
}

func (s *Stack) Push(v Value) {
	s.Stack[s.StackTop] = v
	s.StackTop++
}

func (s *Stack) Dup() {
	s.Push(s.Stack[s.StackTop-1])
}

func (v *VM) ParseValue(op OpCode) Value {
	var value Value
	offset := opcodePushLengths[op]
	data := v.Chunk.Bytecode[v.Ip : v.Ip+offset]
	copy(value[32-offset:], data)

	v.Ip += offset
	return value
}

func (v *VM) BinaryOp(op OpCode) error {
	bValue := v.Stack.Pop()
	aValue := v.Stack.Pop()

	aInt := new(big.Int).SetBytes(aValue[:])
	bInt := new(big.Int).SetBytes(bValue[:])

	resultInt := new(big.Int)
	switch op {
	case OP_ADD:
		resultInt.Add(aInt, bInt)
	case OP_SUB:
		resultInt.Sub(aInt, bInt)
	case OP_MULT:
		resultInt.Mul(aInt, bInt)
	case OP_DIV:
		if bInt.Sign() == 0 {
			resultInt.SetInt64(0)
		} else {
			resultInt.Div(aInt, bInt)
		}
	}
	resultValue := bigIntToValue(resultInt)
	v.Stack.Push(resultValue)

	return nil
}

func bigIntToValue(i *big.Int) Value {
	var v Value
	bytes := i.Bytes()
	copy(v[32-len(bytes):], bytes)
	return v
}

func (v *VM) Consume() byte {
	defer func() { v.Ip++ }()
	return v.Peek(0)
}

func (v *VM) Peek(offset int) byte {
	if v.Ip+offset >= len(v.Chunk.Bytecode) {
		return byte(OP_END)
	}
	return v.Chunk.Bytecode[v.Ip+offset]
}

func (v *VM) Run() error {
	for {
		op := OpCode(v.Consume())
		switch op {
		case OP_PUSH1, OP_PUSH4, OP_PUSH8:
			fmt.Println("OP_PUSH")
			value := v.ParseValue(op)
			v.Stack.Push(value)
		case OP_POP:
			fmt.Println("OP_POP")
			v.Stack.Pop()
		case OP_DUP1:
			fmt.Println("OP_DUP")
			v.Stack.Dup()
		case OP_ADD:
			fmt.Println("OP_ADD")
			err := v.BinaryOp(OP_ADD)
			if err != nil {
				return err
			}
		case OP_SUB:
			fmt.Println("OP_SUB")
			err := v.BinaryOp(OP_SUB)
			if err != nil {
				return err
			}
		case OP_MULT:
			fmt.Println("OP_MULT")
			err := v.BinaryOp(OP_MULT)
			if err != nil {
				return err
			}
		case OP_DIV:
			fmt.Println("OP_DIV")
			err := v.BinaryOp(OP_DIV)
			if err != nil {
				return err
			}
		case MSTORE:
		case MLOAD:
		case SSTORE:
		case SLOAD:
		case OP_PRINT:
			fmt.Println("OP_PRINT")
			v.PrintValue(v.Stack.Pop())
		case OP_END:
			fmt.Println("OP_END")
			return nil
		}
	}
}

func (v *VM) PrintValue(val Value) {
	valInt := new(big.Int).SetBytes(val[:])
	fmt.Println(valInt.String())
}
