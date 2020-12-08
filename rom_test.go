package main

import (
	"testing"
	"bytes"
	"github.com/stretchr/testify/assert"
)

func TestReadROM(t *testing.T) {
	buf := bytes.Buffer{}
	buf.Write([]byte("test"))

	result := ReadROM(&buf, 4)

	assert.Equal(t, result, []byte("test"))
}
