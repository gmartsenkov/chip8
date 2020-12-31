package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	path := flag.String("rom", "", "Path to the chip8 rom")
	flag.Parse()

	if *path == "" {
		fmt.Println("Provide path to rom.\nExample: chip8 --rom ./breakout.ch8")
		os.Exit(1)
	}

	rom, err := os.Open(*path)

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
