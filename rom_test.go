package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadROM(t *testing.T) {
	buf := bytes.Buffer{}
	buf.Write([]uint8("test"))

	result := ReadROM(&buf, 4)

	assert.Equal(t, result, []byte("test"))
}
