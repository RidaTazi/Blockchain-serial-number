package main

import (
	//"encoding/hex"
	//"fmt"
	"testing"

	//"github.com/stretchr/testify/assert"
)


func TestOutput(t *testing.T) {
	output := NewTXOutput("30023002", "1DqtcYLfCT37hDzDF5zAXDxcFvRscnsdRE", "0")
	t.Log(output)
}
