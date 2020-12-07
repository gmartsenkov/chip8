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

	ReadROM(rom, int(romInfo.Size()))
}
