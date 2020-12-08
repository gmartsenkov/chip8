package main

import (
	"bufio"
	"io"
)

func ReadROM(rom io.Reader, size int) []byte {
	reader := bufio.NewReader(rom)
	buffer := make([]byte, size)

	for {
		_, err := reader.Read(buffer)

		if err != nil {
			if err == io.EOF {
				break
			}
		}
	}

	return buffer
}
