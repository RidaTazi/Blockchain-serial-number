package main

import (
	"fmt"
)

func (cli *CLI) getSerialNumber(serialNumber, salt string, nodeID string) {
	bc := NewBlockchain(nodeID)
	defer bc.db.Close()

	hash := HashSerialNumber(serialNumber, salt)

	outputs, txIDs := bc.FindSerialNumberHash(hash)
	
	for index, output := range outputs {
		fmt.Printf("============ Transaction ============\n")
		fmt.Printf("txid: %x\n", txIDs[index])
		fmt.Printf("Serial Number Hash: %d\n", output.SerialNumberHash)
		fmt.Printf("Script (PubKey hash of recipient): %x\n", output.PubKeyHash)
		fmt.Printf("\n")
	}
}
