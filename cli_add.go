package main

import (
	"fmt"
	"log"
)

func (cli *CLI) createNewSerialNumber(to string, serialNumber, salt string, nodeID string, mineNow bool) {
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	tx := NewSerialNumberTX(to, serialNumber, salt)

	if mineNow {
		bc := NewBlockchain(nodeID)
		UTXOSet := UTXOSet{bc}
		defer bc.db.Close()
		txs := []*Transaction{tx}
		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		sendTx(knownNodes[0], tx)
	}

	fmt.Println("Success!")
}
