package main

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
	V [16]uint8 // 16 Registers (V0 to VF)
	Stack [16]uint16

	PC uint16 // Program counter
	SP uint8 // Stack pointer
	I uint16 // Index register
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

func (vm *VM) LoadProgram(program []byte) {
	for i, v := range program {
		vm.Memory[i+512] = v
	}
}
