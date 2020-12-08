package main

import (
	"fmt"
)

type UnknownOpCode struct {
	OpCode uint16
}

func (uo *UnknownOpCode) Error() string {
	return "Unknown OpCode: " + fmt.Sprintf("%X", uo.OpCode)
}

type VM struct {
	// The 4096 bytes of memory.
	//
	// Memory Map:
	// +---------------+= 0xFFF (4095) End of Chip-8 RAM
	// |               |
	// |               |
	// |               |
	// |               |
	// |               |
	// | 0x200 to 0xFFF|
	// |     Chip-8    |
	// | Program / Data|
	// |     Space     |
	// |               |
	// |               |
	// |               |
	// +- - - - - - - -+= 0x600 (1536) Start of ETI 660 Chip-8 programs
	// |               |
	// |               |
	// |               |
	// +---------------+= 0x200 (512) Start of most Chip-8 programs
	// | 0x000 to 0x1FF|
	// | Reserved for  |
	// |  interpreter  |
	// +---------------+= 0x000 (0) Start of Chip-8 RAM

	Memory [4095]byte
	V      [16]uint8 // 16 Registers (V0 to VF)
	Stack  [16]uint16

	PC uint16 // Program counter
	SP uint8  // Stack pointer
	I  uint16 // Index register
}

func InitVM() VM {
	instance := VM{
		PC: 0x200,
	}

	for i, v := range Fonts {
		instance.Memory[i] = v
	}

	return instance
}

func (vm *VM) ExecOp(op uint16) error {
	switch op & 0xF000 {
	case 0x0000: // SYS addr
		switch op {
		case 0x00E0: // CLS
			vm.PC += 2
			break
		case 0x00EE: // RET
			vm.PC = vm.Stack[vm.SP]
			vm.SP--
			vm.PC += 2
			break
		default:
			return &UnknownOpCode{OpCode: op}
		}

		break
	case 0x1000: // Jump address
		vm.PC = op & 0x0FFF
		break
	}
	return nil
}

func (vm *VM) LoadProgram(program []byte) {
	for i, v := range program {
		vm.Memory[i+512] = v
	}
}
