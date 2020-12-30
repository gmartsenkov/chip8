package main

import (
	"log"
	"os"
)

func main() {
	rom, err := os.Open("./roms/maze_demo.ch8")

	if err != nil {
		log.Fatal(err)
	}

	defer rom.Close()

	romInfo, err := rom.Stat()

	if err != nil {
		log.Fatal(err)
	}

	program := ReadROM(rom, int(romInfo.Size()))

	screen := Screen{}
	screen.Init()
	defer screen.Close()

	vm := InitVM()

	vm.SetScreen(&screen)
	vm.LoadProgram(program)
}
