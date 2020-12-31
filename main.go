package main

import (
	"log"
	"os"
)

func main() {
	rom, err := os.Open("./roms/breakout.ch8")

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

	logFile, err := os.OpenFile("chip8.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		os.Exit(1)
	}

	vm.Logger.SetOutput(logFile)

	go vm.EventListener()

	err = vm.Start()

	if err != nil {
		os.Exit(1)
	}
}
