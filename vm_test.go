package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	randByte = func() byte {
		return 0x01
	}
}

func TestInitVM(t *testing.T) {
	vm := InitVM()

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, len(vm.Memory), 4095)
}

func TestLoadProgram(t *testing.T) {
	vm := InitVM()
	program := []byte("program")

	assert.Equal(t, vm.Memory[512:519], []byte{0, 0, 0, 0, 0, 0, 0})

	vm.LoadProgram(program)

	assert.Equal(t, vm.Memory[512:519], []byte("program"))
}

func TestDecodeOpCode(t *testing.T) {
	vm := InitVM()
	program := []byte{0x10, 0x20}

	assert.Equal(t, vm.Memory[512:514], []byte{0, 0})
	vm.LoadProgram(program)
	assert.Equal(t, vm.Memory[512:514], []byte{0x10, 0x20})

	assert.Equal(t, vm.decodeOpCode(), uint16(0x1020))
}

func TestStep(t *testing.T) {
	vm := InitVM()
	vm.DT = 5
	vm.ST = 15
	program := []byte{0x00, 0xE0, 0x00, 0x00, 0x00, 0xE0}
	screen := Screen{Pixels: [2048]byte{1, 2}}

	vm.LoadProgram(program)
	vm.SetScreen(&screen)

	assert.Equal(t, screen.Pixels[0], uint8(1))
	assert.Equal(t, screen.Pixels[1], uint8(2))
	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.DT, uint8(5))
	assert.Equal(t, vm.ST, uint8(15))

	err := vm.Step()
	assert.Nil(t, err)

	assert.Equal(t, screen.Pixels[0], uint8(0))
	assert.Equal(t, screen.Pixels[1], uint8(0))
	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.DT, uint8(4))
	assert.Equal(t, vm.ST, uint8(14))

	err = vm.Step()
	assert.Equal(t, err, &UnknownOpCode{OpCode: 0x0})

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.DT, uint8(4))
	assert.Equal(t, vm.ST, uint8(14))

	vm.DT = 0
	vm.ST = 0
	vm.PC = 0x204

	err = vm.Step()
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x206))
	assert.Equal(t, vm.DT, uint8(0))
	assert.Equal(t, vm.ST, uint8(0))
}

// CLS
func TestExecOpCLS(t *testing.T) {
	vm := InitVM()
	screen := Screen{Pixels: [2048]byte{1, 2}}
	vm.SetScreen(&screen)

	assert.Equal(t, screen.Pixels[0], uint8(1))
	assert.Equal(t, screen.Pixels[1], uint8(2))
	assert.Equal(t, vm.PC, uint16(0x200))

	err := vm.ExecOp(0x00E0)
	assert.Nil(t, err)

	assert.Equal(t, screen.Pixels[0], uint8(0))
	assert.Equal(t, screen.Pixels[1], uint8(0))
	assert.Equal(t, vm.PC, uint16(0x202))
}

// RET
func TestExecOpRET(t *testing.T) {
	vm := InitVM()
	vm.SP = 2
	vm.Stack[2] = 0x300
	assert.Equal(t, vm.PC, uint16(0x200))

	err := vm.ExecOp(0x00EE)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x302))
}

// Invalid SYS address
func TestExecOpInvalidSYS(t *testing.T) {
	vm := InitVM()

	err := vm.ExecOp(0x00E1)
	assert.Equal(t, err, &UnknownOpCode{OpCode: 0x00E1})
}

// JP address
func TestExecOpJPAddr(t *testing.T) {
	vm := InitVM()

	assert.Equal(t, vm.PC, uint16(0x200))

	err := vm.ExecOp(0x1234)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x234))
}

// CALL address
func TestExecOpCallAddr(t *testing.T) {
	vm := InitVM()

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.SP, uint8(0x0))
	assert.Equal(t, vm.Stack, [16]uint16{})

	err := vm.ExecOp(0x2234)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x234))
	assert.Equal(t, vm.SP, uint8(0x1))
	assert.Equal(
		t,
		vm.Stack,
		[16]uint16{0x0, 0x200, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	)
}

// SE Vx
func TestExecOpSEVx(t *testing.T) {
	vm := InitVM()
	vm.V[4] = 0xFF

	assert.Equal(t, vm.PC, uint16(0x200))

	err := vm.ExecOp(0x3456)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))

	err = vm.ExecOp(0x34FF)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x206))
}

// SNE Vx
func TestExecOpSNEVx(t *testing.T) {
	vm := InitVM()
	vm.V[4] = 0xFF

	assert.Equal(t, vm.PC, uint16(0x200))

	err := vm.ExecOp(0x4456)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x204))

	err = vm.ExecOp(0x44FF)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x206))
}

// SE Vx, VY
func TestExecOpSEVxVy(t *testing.T) {
	vm := InitVM()
	vm.V[4] = 0x4
	vm.V[3] = 0x4
	vm.V[2] = 0x3

	assert.Equal(t, vm.PC, uint16(0x200))

	err := vm.ExecOp(0x5430)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x204))

	err = vm.ExecOp(0x5420)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x206))

	err = vm.ExecOp(0x5431)
	assert.Equal(t, err, &UnknownOpCode{OpCode: 0x5431})
}

// LD Vx, byte
func TestExecOpLDVx(t *testing.T) {
	vm := InitVM()
	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x0))

	err := vm.ExecOp(0x6235)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x35))
}

// ADD Vx, byte
func TestExecOpAddVx(t *testing.T) {
	vm := InitVM()
	vm.V[2] = 0x10
	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x10))

	err := vm.ExecOp(0x7210)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x20))
}

// LD Vx, Vy
func TestExecOpLDVxVy(t *testing.T) {
	vm := InitVM()
	vm.V[2] = 0x05
	vm.V[3] = 0x10

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x05))
	assert.Equal(t, vm.V[3], uint8(0x10))

	err := vm.ExecOp(0x8230)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x10))
	assert.Equal(t, vm.V[3], uint8(0x10))
}

// OR Vx, Vy
func TestExecOpORVxVy(t *testing.T) {
	vm := InitVM()
	vm.V[2] = 0x10
	vm.V[3] = 0x01

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x10))
	assert.Equal(t, vm.V[3], uint8(0x01))

	err := vm.ExecOp(0x8231)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x11))
	assert.Equal(t, vm.V[3], uint8(0x01))
}

// AND Vx, Vy
func TestExecOpANDVxVy(t *testing.T) {
	vm := InitVM()
	vm.V[2] = 0x15
	vm.V[3] = 0x05

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x15))
	assert.Equal(t, vm.V[3], uint8(0x05))

	err := vm.ExecOp(0x8232)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x05))
	assert.Equal(t, vm.V[3], uint8(0x05))
}

// XOR Vx, Vy
func TestExecOpXORVxVy(t *testing.T) {
	vm := InitVM()
	vm.V[2] = 0x15
	vm.V[3] = 0x05

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x15))
	assert.Equal(t, vm.V[3], uint8(0x05))

	err := vm.ExecOp(0x8233)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x10))
	assert.Equal(t, vm.V[3], uint8(0x05))
}

// ADD Vx, Vy
func TestExecOpAddVxVy(t *testing.T) {
	vm := InitVM()
	vm.V[2] = 0x15
	vm.V[3] = 0x05
	vm.V[4] = 0xFF

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x15))
	assert.Equal(t, vm.V[3], uint8(0x05))
	assert.Equal(t, vm.V[4], uint8(0xFF))
	assert.Equal(t, vm.V[0xF], uint8(0))

	err := vm.ExecOp(0x8234)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x1A))
	assert.Equal(t, vm.V[3], uint8(0x05))
	assert.Equal(t, vm.V[4], uint8(0xFF))

	err = vm.ExecOp(0x8244)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x204))
	assert.Equal(t, vm.V[2], uint8(0x19))
	assert.Equal(t, vm.V[3], uint8(0x05))
	assert.Equal(t, vm.V[4], uint8(0xFF))
	assert.Equal(t, vm.V[0xF], uint8(1))
}

// SUB Vx, Vy
func TestExecOpSUBVxVy(t *testing.T) {
	vm := InitVM()
	vm.V[2] = 0x15
	vm.V[3] = 0x05
	vm.V[4] = 0xFF

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x15))
	assert.Equal(t, vm.V[3], uint8(0x05))
	assert.Equal(t, vm.V[4], uint8(0xFF))
	assert.Equal(t, vm.V[0xF], uint8(0))

	err := vm.ExecOp(0x8235)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x10))
	assert.Equal(t, vm.V[3], uint8(0x05))
	assert.Equal(t, vm.V[4], uint8(0xFF))
	assert.Equal(t, vm.V[0xF], uint8(1))

	err = vm.ExecOp(0x8245)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x204))
	assert.Equal(t, vm.V[2], uint8(0x11))
	assert.Equal(t, vm.V[3], uint8(0x05))
	assert.Equal(t, vm.V[4], uint8(0xFF))
	assert.Equal(t, vm.V[0xF], uint8(0))
}

// SHR Vx, {, Vy}
func TestExecOpSHRVxVy(t *testing.T) {
	vm := InitVM()
	vm.V[2] = 0x11
	vm.V[4] = 0x00

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x11))
	assert.Equal(t, vm.V[4], uint8(0x00))
	assert.Equal(t, vm.V[0xF], uint8(0))

	err := vm.ExecOp(0x8236)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x08))
	assert.Equal(t, vm.V[4], uint8(0x00))
	assert.Equal(t, vm.V[0xF], uint8(1))

	err = vm.ExecOp(0x8436)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x204))
	assert.Equal(t, vm.V[2], uint8(0x08))
	assert.Equal(t, vm.V[4], uint8(0x00))
	assert.Equal(t, vm.V[0xF], uint8(0))
}

// SUBN Vx, Vy
func TestExecOpSUBNVxVy(t *testing.T) {

	vm := InitVM()
	vm.V[2] = 0x05
	vm.V[3] = 0x08
	vm.V[4] = 0x2

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x05))
	assert.Equal(t, vm.V[3], uint8(0x08))
	assert.Equal(t, vm.V[4], uint8(0x2))
	assert.Equal(t, vm.V[0xF], uint8(0))

	err := vm.ExecOp(0x8237)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x3))
	assert.Equal(t, vm.V[3], uint8(0x08))
	assert.Equal(t, vm.V[4], uint8(0x2))
	assert.Equal(t, vm.V[0xF], uint8(1))

	err = vm.ExecOp(0x8247)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x204))
	assert.Equal(t, vm.V[2], uint8(0xff))
	assert.Equal(t, vm.V[3], uint8(0x08))
	assert.Equal(t, vm.V[4], uint8(0x2))
	assert.Equal(t, vm.V[0xF], uint8(0))
}

// SHL Vx, {, Vy}
func TestExecOpSHLVxVy(t *testing.T) {
	vm := InitVM()
	vm.V[2] = 0x11
	vm.V[4] = 0x80

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x11))
	assert.Equal(t, vm.V[4], uint8(0x80))
	assert.Equal(t, vm.V[0xF], uint8(0))

	err := vm.ExecOp(0x823E)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x22))
	assert.Equal(t, vm.V[4], uint8(0x80))
	assert.Equal(t, vm.V[0xF], uint8(0))

	err = vm.ExecOp(0x843E)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x204))
	assert.Equal(t, vm.V[2], uint8(0x22))
	assert.Equal(t, vm.V[4], uint8(0x0))
	assert.Equal(t, vm.V[0xF], uint8(1))
}

// SNE Vx, Vy
func TestExecOpSNEVxVy(t *testing.T) {
	vm := InitVM()
	vm.V[2] = 0x11
	vm.V[3] = 0x11
	vm.V[4] = 0x80

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x11))
	assert.Equal(t, vm.V[3], uint8(0x11))
	assert.Equal(t, vm.V[4], uint8(0x80))
	assert.Equal(t, vm.V[0xF], uint8(0))

	err := vm.ExecOp(0x9230)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x11))
	assert.Equal(t, vm.V[3], uint8(0x11))
	assert.Equal(t, vm.V[4], uint8(0x80))
	assert.Equal(t, vm.V[0xF], uint8(0))

	err = vm.ExecOp(0x9240)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x206))
	assert.Equal(t, vm.V[2], uint8(0x11))
	assert.Equal(t, vm.V[3], uint8(0x11))
	assert.Equal(t, vm.V[4], uint8(0x80))
	assert.Equal(t, vm.V[0xF], uint8(0))

	err = vm.ExecOp(0x9241)
	assert.Equal(t, err, &UnknownOpCode{OpCode: 0x9241})
}

// LD I, addr
func TestExecOpLDI(t *testing.T) {
	vm := InitVM()

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.I, uint16(0x0))

	err := vm.ExecOp(0xA234)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.I, uint16(0x234))
}

// JP v0, addr
func TestExecOpJPV0(t *testing.T) {
	vm := InitVM()
	vm.V[0] = 0x01
	assert.Equal(t, vm.PC, uint16(0x200))

	err := vm.ExecOp(0xB123)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x124))
}

// RND Vx, byte
func TestExecOpRNDVx(t *testing.T) {
	vm := InitVM()
	vm.V[2] = 0x15

	assert.Equal(t, vm.PC, uint16(0x200))
	assert.Equal(t, vm.V[2], uint8(0x15))

	err := vm.ExecOp(0xC223)
	assert.Nil(t, err)

	assert.Equal(t, vm.PC, uint16(0x202))
	assert.Equal(t, vm.V[2], uint8(0x24))
}
