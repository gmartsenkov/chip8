package main

import (
	"fmt"
	"os"

	"github.com/nsf/termbox-go"
)

const (
	width  = 64
	height = 32
)

// Contains the pixels on screen and implements screen render related functions
type Screen struct {
	Pixels [width * height]byte
}

func (screen *Screen) Init() {
	err := termbox.Init()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (screen *Screen) Close() {
	termbox.Close()
}

func (screen *Screen) Render() {
	for row := 0; row < height; row++ {
		for pixel := 0; pixel < width; pixel++ {
			v := ' '

			if v == 0x01 {
				v = '█'
			}
			termbox.SetCell(row, pixel, v, termbox.ColorGreen, termbox.ColorBlack)
		}
	}

	termbox.Flush()
}
