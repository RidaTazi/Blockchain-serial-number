package main

import (
	"fmt"
	"log"
)

func (cli *CLI) startNode(nodeAddress string, minerAddress, apiAddress string) {
	
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Println("Mining is on: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	
	StartServer(nodeAddress, minerAddress, apiAddress)
}
