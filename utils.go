package main

import (
	"bytes"
	"encoding/binary"
	"crypto/sha256"
	"math"
	"strings"
	"log"
)

// Need to test this function
func Float64ToByte(float float64) []byte {
    bits := math.Float64bits(float)
    bytes := make([]byte, 8)
    binary.LittleEndian.PutUint64(bytes, bits)

    return bytes
}

func ParseNodeID(nodeAddress string) string {
	a := strings.Split(nodeAddress, ":")
	if len(a) != 2 {
		log.Panic("Invalid node address")
	}
	return a[1]
}

func SliceContainsString(list []string, str string) bool {
 	for _, v := range list {
 		if strings.Compare(str, v) == 0 {
 			return true
 		}
 	}
 	return false
}

func HashSerialNumber(serialNumber, salt string) []byte {
	payload := append([]byte(serialNumber), salt...)

	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:]
}

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
