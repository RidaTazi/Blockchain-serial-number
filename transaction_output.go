package main

import (
	"bytes"
	"fmt"
	"strings"
	"encoding/gob"
	"crypto/sha256"
	"log"
)

type TXOutput struct {
	SerialNumberHash	[]byte  // hash(serial number + salt)
	PubKeyHash  		[]byte  // hash(pubKey) of the recipient
}

func (out TXOutput) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf(" === Output ==="))
	lines = append(lines, fmt.Sprintf("       Serial Number Hash:  %x", out.SerialNumberHash))
	lines = append(lines, fmt.Sprintf("       Script: %x", out.PubKeyHash))

	return strings.Join(lines, "\n")
}

// Lock signs the output
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func (out *TXOutput) UpdateSerialNumberHash(serialNumber, salt string) {
	payload := append([]byte(serialNumber), salt...)

	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	out.SerialNumberHash = secondSHA[:]
}

// NewTXOutput create a new TXOutput
func NewTXOutput(serialNumber, recipient_addr string, salt string) *TXOutput {
	txo := &TXOutput{}
	txo.UpdateSerialNumberHash(serialNumber, salt)
	txo.Lock([]byte(recipient_addr))

	return txo
}

// TXOutputs collects TXOutput
type TXOutputs struct {
	Outputs []TXOutput
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}
