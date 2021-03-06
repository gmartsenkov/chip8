package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

const (
	clockSpeed = time.Duration(120)
	resetKeySpeed = time.Duration(6)
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

	Screen *Screen
	Keypad Keypad
	Logger log.Logger

	Clock          <-chan time.Time // Timer
	ResetKeysClock <-chan time.Time // Reset pressed keys timer
	Render         chan int         // Render
	Event          chan byte        // Key press

	DT uint8 // Delay Timer
	ST uint8 // Sound Timer
}

func InitVM() VM {
	instance := VM{
		PC:    0x200,
		Clock: time.Tick(time.Second / clockSpeed),
		ResetKeysClock: time.Tick(time.Second / resetKeySpeed),
	}
	instance.Render = make(chan int, 5)
	instance.Event = make(chan byte, 10)

	for i, v := range Fonts {
		instance.Memory[i] = v
	}

	return instance
}

func (vm *VM) SetScreen(screen *Screen) {
	vm.Screen = screen
}

func (vm *VM) Step() error {
	op := vm.decodeOpCode()

	err := vm.ExecOp(op)
	if err != nil {
		return err
	}

	if vm.DT > 0 {
		vm.DT -= 1
	}
	if vm.ST > 0 {
		vm.ST -= 1
	}

	return nil
}

func (vm *VM) Start() error {
	for {
		select {
		case event := <-vm.Event:
			if event == 0x00 {
				return nil
			}
			vm.Keypad.PressKey(event)
		case <- vm.ResetKeysClock:
			vm.Keypad.Reset()
		case <-vm.Clock:
			if err := vm.Step(); err != nil {
				return err
			}
		case <-vm.Render:
			vm.Screen.Render()
		}
	}
}

func (vm *VM) EventListener() {
	for {
		vm.Event <- pollEvent()
	}
}

func (vm *VM) ExecOp(op uint16) error {
	switch op & 0xF000 {
	case 0x0000: // SYS addr
		switch op {
		case 0x00E0: // CLS
			vm.Screen.Clear()

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
	case 0xD000: // DRW Vx, Vy, nibble
		x := vm.V[op&0x0F00>>8]
		y := vm.V[op&0x00F0>>4]
		nibble := op & 0x000F

		collision := vm.Screen.WriteSprite(vm.Memory[vm.I:vm.I+nibble], x, y)

		if collision {
			vm.V[0xF] = 1
		} else {
			vm.V[0xF] = 0
		}

		vm.Render <- 0

		vm.PC += 2
		break
	case 0xE000:
		x := vm.V[op&0x0F00>>8]

		switch op & 0x00FF {
		case 0x009E: // Ex9E - SKP Vx
			if vm.Keypad.CheckPressed(x) {
				vm.PC += 2
			}

			vm.PC += 2
			break
		case 0x00A1: // ExA1 - SKPN Vx
			if !vm.Keypad.CheckPressed(x) {
				vm.PC += 2
			}

			vm.PC += 2
			break
		default:
			return &UnknownOpCode{OpCode: op}
		}

		break
	case 0xF000:
		x := op & 0x0F00 >> 8

		switch op & 0x00FF {
		case 0x0007: // LD Vx, DT
			vm.V[x] = vm.DT
			vm.PC += 2
			break
		case 0x000A: // LD Vx, K
			vm.V[x] = fetchKey()
			vm.PC += 2
			break
		case 0x0015: // LD DT, Vx
			vm.DT = vm.V[x]
			vm.PC += 2
			break
		case 0x0018: // LD ST, Vx
			vm.ST = vm.V[x]
			vm.PC += 2
			break
		case 0x001E: // ADD I, Vx
			vm.I += uint16(vm.V[x])
			vm.PC += 2
			break
		case 0x0029: // LD F, Vx
			vm.I = uint16(vm.V[x]) * 5
			vm.PC += 2
			break
		case 0x0033: // LD B, Vx
			vm.Memory[vm.I] = vm.V[x] / 100
			vm.Memory[vm.I+1] = (vm.V[x] / 10) % 10
			vm.Memory[vm.I+2] = (vm.V[x] % 100) % 10
			vm.PC += 2
			break
		case 0x0055: // LD [I], Vx
			for i := 0; uint16(i) <= x; i++ {
				vm.Memory[vm.I+uint16(i)] = vm.V[i]
			}

			vm.PC += 2
			break
		case 0x0065: // LD Vx, [I]
			for i := 0; uint16(i) <= x; i++ {
				vm.V[i] = vm.Memory[vm.I+uint16(i)]
			}

			vm.PC += 2
			break
		default:
			return &UnknownOpCode{OpCode: op}
		}

		break
	}
	return nil
}

func (vm *VM) decodeOpCode() uint16 {
	return uint16(vm.Memory[vm.PC])<<8 | uint16(vm.Memory[vm.PC+1])
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
