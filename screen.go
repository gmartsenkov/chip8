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

func (screen *Screen) Clear() {
	for i := range screen.Pixels {
		screen.Pixels[i] = 0
	}
}

func (screen *Screen) WriteSprite(sprite []byte, x, y byte) bool {
	collision := false
	spriteHeight := len(sprite)

	// fmt.Printf("Sprite: %b, X: %d, Y: %d\n", sprite, x, y)
	for yline := 0; yline < spriteHeight; yline++ {
		pixel := sprite[yline]
		// fmt.Printf("Pixel: %x\n", pixel)

		for xline := 0; xline < 8; xline++ {
			if (pixel & (0x80 >> xline)) != 0 {
				position := (((int(x) + xline) % width) + ((int(y) + yline) * width)) % (width * height)

				// fmt.Printf("Pos x: %d, Pos y: %d, Real: %d\n", int(x)+xline, int(y)+yline, position)
				if screen.Pixels[position] == 0x01 {
					collision = true
				}

				screen.Pixels[position] ^= 1
			}
		}
	}

	return collision
}

func (screen *Screen) Render() {
	for row := 0; row < height; row++ {
		for pixel := 0; pixel < width; pixel++ {
			v := ' '
			coord := row*width + pixel

			if screen.Pixels[coord] == 0x01 {
				v = '█'
			}
			termbox.SetCell(pixel, row, v, termbox.ColorGreen, termbox.ColorBlack)
		}
	}

	termbox.Flush()
}
