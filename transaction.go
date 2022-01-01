package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"strings"

	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID   []byte      // Txid
	Vin  []TXInput   // list of inputs
	Vout []TXOutput  // list of outputs
}

func (tx Transaction) IsNewSerialNumberTX() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// NewCoinbaseTX creates a new coinbase transaction
func NewSerialNumberTX(recipient_addr, serialNumber string, salt string) *Transaction {
	// func NewTXOutput(serialNumber string, address string, salt string) *TXOutput
	txout := *NewTXOutput(serialNumber, recipient_addr, salt)
	txin := TXInput{[]byte{}, -1, nil, []byte(txout.SerialNumberHash)}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.ID = tx.Hash()

	return &tx
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(wallet *Wallet, address, serialNumber string, salt string, UTXOSet *UTXOSet) (*Transaction, error) {
	var inputs []TXInput
	var outputs []TXOutput

	pubKeyHash := HashPubKey(wallet.PublicKey)
	serialNumberHash := HashSerialNumber(serialNumber, salt)

	validOutput := *UTXOSet.FindValidOutput(pubKeyHash, serialNumberHash)
	if validOutput.Txid == nil {
		return nil, fmt.Errorf("No valid output found. Make sure serial number and salt are correct.")
	}

	// built input
	txID, err := hex.DecodeString(string(validOutput.Txid[:]))
	if err != nil {log.Panic(err)}
	input := TXInput{txID, validOutput.Vout, nil, wallet.PublicKey}
	inputs = append(inputs, input)

	// build output
	// func NewTXOutput(serialNumber string, address string, salt string) *TXOutput
	outputs = append(outputs, *NewTXOutput(serialNumber, address, salt))

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)

	return &tx, nil
}

/*
// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(wallet *Wallet, addresses, serialNumbers []string, salt string, UTXOSet *UTXOSet) (*Transaction, []string) {
	var inputs []TXInput
	var outputs []TXOutput

	if len(addresses) != len(serialNumbers) {
		log.Panic("ERROR: addresses and serial numbers must pair up")
	}

	pubKeyHash := HashPubKey(wallet.PublicKey)
	
	var serialNumberHashes map[string]int
	for i, serialNumber := range serialNumbers {
		hash := HashSerialNumber(serialNumber, salt)
		serialNumberHashes[string(hash[:])] = i
	}
	// []int, []ValidOutput
	//invalidIndices, validOutputs := UTXOSet.FindValidSerialNumbers(pubKeyHash, serialNumberHashes)
	_, validOutputs := UTXOSet.FindValidSerialNumbers(pubKeyHash, serialNumberHashes)

	for _, validOutput := range validOutputs {
		// built input
		txID, err := hex.DecodeString(validOutput.Txid)
		if err != nil {log.Panic(err)}
		input := TXInput{txID, validOutput.Vout, nil, wallet.PublicKey}
		inputs = append(inputs, input)

		// build output
		serialNumber := serialNumbers[validOutput.Index]
		address := addresses[validOutput.Index]
		// func NewTXOutput(serialNumber, address string, salt string) *TXOutput
		outputs = append(outputs, *NewTXOutput(serialNumber, address, salt))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)

	return &tx, nil
}*/

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsNewSerialNumberTX() {
		return
	}

	// Check previous transactions are correct
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		vin.Signature = nil
		vin.PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
		txCopy.Vin[inID].PubKey = nil
	}
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.SerialNumberHash, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// Verify verifies signatures of Transaction inputs
// TODO: how to verify new serial number TX?
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsNewSerialNumberTX() {
		return true
	}

	// Check previous transactions are correct
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.Vin[inID].PubKey = nil
	}

	return true
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x ---", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	//type TXOutput struct {
	//	SerialNumberHash	[]byte  // hash(serial number + salt)
	//	PubKeyHash  		[]byte  // hash(pubKey) of the recipient, not pubKey.
	//}
	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Serial Number Hash:  %x", output.SerialNumberHash))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// DeserializeTransaction deserializes a transaction
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}
