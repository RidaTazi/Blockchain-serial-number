package main

import (
	"fmt"
)

func (cli *CLI) send(from string, to, serialNumber string, salt, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		fmt.Errorf("Error: Sender address is not valid\n")
		return
	}
	if !ValidateAddress(to) {
		fmt.Errorf("Error: Recipient address is not valid\n")
		return
	}

	bc := NewBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		fmt.Errorf("Error: %s\n", err)
		return
	}

	wallet := wallets.GetWallet(from)
	if wallet == nil {
		fmt.Errorf("You do not own the address: %s", from)
		return
	}

	// later may be modified to transfer labels in patch
	tx, err := NewUTXOTransaction(wallet, to, serialNumber, salt, &UTXOSet)

	if err != nil {
		fmt.Println(err)
		return
	}

	if mineNow {
		txs := []*Transaction{tx}
		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		sendTx(knownNodes[0], tx)
	}

}


