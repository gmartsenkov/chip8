package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

// CLS
func TestExecOpCLS(t *testing.T) {
	vm := InitVM()
	assert.Equal(t, vm.PC, uint16(0x200))

	err := vm.ExecOp(0x00E0)
	assert.Nil(t, err)

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
