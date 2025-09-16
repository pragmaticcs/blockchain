package vm

import (
	"errors"
	"fmt"
)

const (
	StackSize = 256
)

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

func (v *VM) ParseValue() Value {
	var value Value
	value.Type = ValueType(v.Chunk.Bytecode[v.Ip])
	v.Consume()    // Consume the type byte
	var offset int // check the correct offset
	switch value.Type {
	case ValueTypeInt:
		offset = ValueIntSize
	case ValueTypeBool:
		offset = ValueBoolSize
	}
	value.Data = v.Chunk.Bytecode[v.Ip : v.Ip+offset]
	v.Ip += offset
	return value
}

func (v *VM) BinaryOp(op OpCode) error {
	a := v.Stack.Pop()
	b := v.Stack.Pop()
	if a.Type != ValueTypeInt || b.Type != ValueTypeInt {
		return errors.New("values must be ints")
	}
	_, aInt, _ := a.VisitAsInt()
	_, bInt, _ := b.VisitAsInt()
	var cInt int32
	switch op {
	case OP_ADD:
		cInt = aInt + bInt
	case OP_MULT:
		cInt = aInt * bInt
	case OP_SUB:
		cInt = aInt - bInt
	case OP_DIV:
		cInt = aInt / bInt
	}
	value, _ := NewValueInt(cInt)
	v.Stack.Push(value)
	return nil
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
		switch OpCode(v.Consume()) {
		case OP_PUSH:
			fmt.Println("OP_PUSH")
			value := v.ParseValue()
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
	switch val.Type {
	case ValueTypeInt:
		_, x, _ := val.VisitAsInt()
		fmt.Println(x)
	case ValueTypeBool:
		_, x, _ := val.VisitAsBool()
		fmt.Println(x)
	}
}
