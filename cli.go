package main

import (
	"flag"
	"fmt"
	"log"

	"os"
)

// CLI responsible for processing command line arguments
type CLI struct{}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
	fmt.Println("  send -from FROM -to TO -data serialNumber -salt salt -mine - Transfer serialNumber from FROM to TO. Mine on the same node, when -mine is set.")
	fmt.Println("  add -to TO -data serialNumber -salt salt -mine - Create a new serialNumber under address TO. Mine on the same node, when -mine is set.")
	fmt.Println("  get -data serialNumber -salt salt - Trace the serial number.")
	fmt.Println("  help - Print usage.")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env. var is not set!")
		os.Exit(1)
	}

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	sendData := sendCmd.String("data", "", "The serial number to send")
	sendSalt := sendCmd.String("salt", "", "The salt")

	addTo := addCmd.String("to", "", "Destination wallet address")
	addData := addCmd.String("data", "", "The serial number to create")
	addSalt := addCmd.String("salt", "", "The salt")
	addMine := addCmd.Bool("mine", false, "Mine immediately on the same node")

	getSerialNumber := getCmd.String("data", "", "The serial number to query")
	getSalt := getCmd.String("salt", "", "The salt")

	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")
	
	switch os.Args[1] {

	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "add":
		err := addCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "get":
		err := getCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress, nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO(nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendData == "" || *sendSalt == "" {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendData, *sendSalt, nodeID, *sendMine)
	}

	if addCmd.Parsed() {
		if *addTo == "" || *addData == "" || *addSalt == "" {
			addCmd.Usage()
			os.Exit(1)
		}

		cli.createNewSerialNumber(*addTo, *addData, *addSalt, nodeID, *addMine)
	}

	if getCmd.Parsed() {
		if *getSerialNumber == "" || *getSalt == "" {
			getCmd.Usage()
			os.Exit(1)
		}
		cli.getSerialNumber(*getSerialNumber, *getSalt, nodeID)
	}

	if startNodeCmd.Parsed() {
		nodeAddress := os.Getenv("NODE_ADDR")
		apiAddress := os.Getenv("API_ADDR")
		if nodeAddress == "" {
			startNodeCmd.Usage()
			os.Exit(1)
		}
		cli.startNode(nodeAddress, *startNodeMiner, apiAddress)
	}
}
