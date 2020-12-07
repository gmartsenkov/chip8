package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
)

func ReadROM(rom io.Reader, size int) []byte {
	reader := bufio.NewReader(rom)
	buffer := make([]byte, size)

	for {
		_, err := reader.Read(buffer)

		if err != nil {
			if err == io.EOF {
				fmt.Println(err)
				break
			}
		}

		fmt.Println("%s", hex.Dump(buffer))
	}

	return buffer
}
