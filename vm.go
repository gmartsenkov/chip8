package main

import (
	"fmt"
	"math/rand"
	"time"
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
	case 0x1000: // 1nnn - Jump address
		vm.PC = op & 0x0FFF
		break
	case 0x2000: // 2nnn - Call addr
		vm.SP++
		vm.Stack[vm.SP] = vm.PC
		vm.PC = op & 0x0FFF
		break
	case 0x3000: // 3xkk - SE Vx - Skip next instruction if Vx = kk.
		x := op & 0x0F00 >> 8
		kk := uint8(op & 0x00FF)

		vm.PC += 2

		if vm.V[x] == kk {
			vm.PC += 2
		}

		break
	case 0x4000: // 4xkk - SNE Vx, byte - Skip next instruction if Vx != kk.
		x := op & 0x0F00 >> 8
		kk := uint8(op & 0x00FF)

		vm.PC += 2

		if vm.V[x] != kk {
			vm.PC += 2
		}

		break
	case 0x5000: // 5xy0 - SE Vx, Vy
		switch op & 0xF00F {
		case 0x5000:
			x := op & 0x0F00 >> 8
			y := op & 0x00F0 >> 4

			vm.PC += 2

			if vm.V[x] == vm.V[y] {
				vm.PC += 2
			}
			break
		default:
			return &UnknownOpCode{OpCode: op}
		}
		break
	case 0x6000: // 6xkk - LD Vx, byte
		x := op & 0x0F00 >> 8
		kk := uint8(op & 0x00FF)

		vm.V[x] = kk

		vm.PC += 2
		break
	case 0x7000: // 7xkk - ADD Vx, byte
		x := op & 0x0F00 >> 8
		kk := uint8(op & 0x00FF)
		vm.V[x] += kk

		vm.PC += 2
		break
	case 0x8000:
		x := op & 0x0F00 >> 8
		y := op & 0x00F0 >> 4

		switch op & 0x000F {
		case 0x0000: // 8xy0 - LD Vx, Vy
			vm.V[x] = vm.V[y]
			vm.PC += 2
			break
		case 0x0001: // 8xy1 - OR Vx, Vy
			vm.V[x] = vm.V[x] | vm.V[y]
			vm.PC += 2
			break
		case 0x0002: // 8xy2 - AND Vx, Vy
			vm.V[x] = vm.V[x] & vm.V[y]
			vm.PC += 2
			break
		case 0x0003: // 8xy3 - XOR Vx, Vy
			vm.V[x] = vm.V[x] ^ vm.V[y]
			vm.PC += 2
			break
		case 0x0004: // 8xy4 - ADD Vx, Vy
			sum := uint16(vm.V[x]) + uint16(vm.V[y])

			var carryFlag byte
			if sum > 255 {
				carryFlag = 1
			}

			vm.V[0xF] = carryFlag
			vm.V[x] = uint8(sum)

			vm.PC += 2

			break
		case 0x0005: // 8xy5 - SUB Vx, Vy
			sum := uint16(vm.V[x]) - uint16(vm.V[y])

			var carryFlag byte
			if vm.V[x] > vm.V[y] {
				carryFlag = 1
			}

			vm.V[0xF] = carryFlag

			vm.V[x] = uint8(sum)

			vm.PC += 2

			break

		case 0x0006: // 8xy6 - SHR Vx {, Vy}
			var carryFlag byte

			if (vm.V[x] & 0x01) == 0x01 {
				carryFlag = 1
			}

			vm.V[0xF] = carryFlag

			vm.V[x] /= 2

			vm.PC += 2

			break
		case 0x0007: // 8xy7 SUBN Vx, Vy
			sum := uint16(vm.V[y]) - uint16(vm.V[x])

			var carryFlag byte

			if vm.V[y] > vm.V[x] {
				carryFlag = 1
			}

			vm.V[0xF] = carryFlag

			vm.V[x] = uint8(sum)
			vm.PC += 2

			break

		case 0x000E: // 8xyE - SHL Vx {, Vy}
			var carryFlag byte

			if (vm.V[x] & 0x80) == 0x80 {
				carryFlag = 1
			}

			vm.V[0xF] = carryFlag

			vm.V[x] *= 2

			vm.PC += 2

			break
		default:
			return &UnknownOpCode{OpCode: op}
		}
		break
	case 0x9000: // 9xy0 SNE Vx, Vy
		x := op & 0x0F00 >> 8
		y := op & 0x00F0 >> 4

		switch op & 0x000F {
		case 0x0000:
			vm.PC += 2

			if vm.V[x] != vm.V[y] {
				vm.PC += 2
			}

			break
		default:
			return &UnknownOpCode{OpCode: op}
		}
	case 0xA000: // Annn - LD I, addr
		vm.I = op & 0x0FFF
		vm.PC += 2
		break
	case 0xB000: // Bnnn - JP v0, addr
		vm.PC = (op & 0x0FFF) + uint16(vm.V[0])
		break
	case 0xC000: // Cxkk - RND Vx, byte
		x := op & 0x0F00 >> 8
		kk := byte(op)

		vm.V[x] = kk + randByte()
		vm.PC += 2
		break
	}
	return nil
}

func (vm *VM) LoadProgram(program []byte) {
	for i, v := range program {
		vm.Memory[i+512] = v
	}
}

// randByte returns a random value between 0 and 255.
var randByte = func() byte {
	return byte(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(255))
}
