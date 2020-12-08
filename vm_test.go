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

	assert.Equal(t, vm.Memory[512:519], []byte{0,0,0,0,0,0,0})

	vm.LoadProgram(program)

	assert.Equal(t, vm.Memory[512:519], []byte("program"))
}
