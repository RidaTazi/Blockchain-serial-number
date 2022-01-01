package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"encoding/json"
	"bufio"
	"os"
	"strings"
	"strconv"
	"time"
	"github.com/fatih/color"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12
const mining_threshold = 2 //start mining after receiving 2 tx
const check_availability_interval = 5 * time.Second

var mutex = &sync.Mutex{}
var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"} // server node IP will be hard-coded
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)
var isClient = false
var isServer = false
var isMiner = false

type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

type getblocks struct {
	AddrFrom string
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

type getJSONReq struct {
	SerialNumber string `json:"serialnumber"`
	Salt 		 string `json:"salt"`
}

type getJSONResp struct {
	Txid 		string
	PubKeyHash 	string
}

/* 	client node: node_address, "", ""
	server node: node_address, "", api_address
	miner node: node_address, miner_address, "" (node_address = miner_address)
*/

func StartServer(node_address string, miner_address string, api_address string) {
	// set global variables
	nodeAddress = node_address
	miningAddress = miner_address

	color.Green("Starting node -> %s\n", nodeAddress)
	if len(api_address) > 0 {
		color.Green("API IP:port -> %s\n", api_address)
	}

	bc := NewBlockchain(ParseNodeID(node_address))

	var wg sync.WaitGroup
	wg.Add(1)
	go startNodeCommunication(node_address, bc)

	// if tracker node
	if knownNodes[0] == nodeAddress {
		isServer = true
		wg.Add(1)
		go launchApiListener(api_address, bc)
	} else {
		//if client or miner, check in with the tracker
		sendVersion(knownNodes[0], bc) 
	}

	// if client node
	if miner_address == "" && api_address == "" {
		isClient = true
		wg.Add(1)
		time.Sleep(500 * time.Millisecond)
		launchClientInterface(bc)
	}

	// if miner node
	if len(miningAddress) > 0 {
		isMiner = true
	}

	wg.Wait()
}

// CLI on client node. Need this because the bolt DB does not permit concurrent access
func launchClientInterface(bc *Blockchain) {
	buf := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		sentence, err := buf.ReadBytes('\n')
		if err != nil {
			fmt.Println(err)
			continue
		}

		line := strings.TrimSuffix(string(sentence), "\n")
		args := strings.Split(line, " ")

		//TODO: error checking on args

		switch args[0] {
		case "add":
			if len(args) != 4 {
				color.Red("Usage: add [recipient addr] [data] [salt]\n")
				break
			}
			to := args[1]
			data := args[2]
			salt := args[3]
			clientAddHandler(to, data, salt, bc)
			
		case "get":
			if len(args) != 3 {
				color.Red("Usage: get [data] [salt]\n")
				break
			}
			data := args[1]
			salt := args[2]
			clientGetHandler(data, salt, bc)
		case "send":
			if len(args) != 5 {
				color.Red("Usage: send [sender addr] [recipient addr] [data] [salt]\n")
				break
			}
			from := args[1]
			to := args[2]
			data := args[3]
			salt := args[4]
			clientSendHandler(from, to, data, salt, bc)
		case "print":
			clientPrintHandler(bc)
		default:
			if args[0] != "" {
				color.Red("Unknown command!")
			}
		}
	}
}

// handle API requests
func launchApiListener(apiAddress string, bc *Blockchain) {
	//serial number, salt -> txid, from pubkey, to pubkey hash
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		req := getJSONReq{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != io.EOF && err != nil {
			panic(err)
		}
		
		resp := getJSONResp{}

		hash := HashSerialNumber(req.SerialNumber, req.Salt)
		outputs, txIDs := bc.FindSerialNumberHash(hash)
		if len(outputs) > 0 {
			resp.Txid = fmt.Sprintf("%s", txIDs[0])
			resp.PubKeyHash = fmt.Sprintf("%x", outputs[0].PubKeyHash)
		} 
		/*
		type getJSONResp struct {
			Txid 		string
			PubKeyHash 	string
		}*/
		respJson, err := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(append(respJson, []byte{'\n'}...))
	})

	log.Fatal(http.ListenAndServe(apiAddress, nil))
	
	/*s := &http.Server{
		Addr:           apiAddress
		Handler:        handler_name,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())*/
}

func clientAddHandler(to string, serialNumber, salt string, bc *Blockchain) {
	if !ValidateAddress(to) {
		log.Println("Error: Recipient address is not valid")
		return
	}

	tx := NewSerialNumberTX(to, serialNumber, salt)

	for _, nodeAddr := range knownNodes {
		sendTx(nodeAddr, tx)
	}

	fmt.Println("Success!")
}

// WARNING: the NewWallets line has volatile dependency
func clientSendHandler(from, to string, serialNumber, salt string, bc *Blockchain) {
	if !ValidateAddress(from) {
		color.Red("Invalid sender address: %s\n", from)
		return
	}
	if !ValidateAddress(to) {
		color.Red("Invalid recipient address: %s\n", to)
		return
	}

	UTXOSet := UTXOSet{bc}

	wallets, err := NewWallets(ParseNodeID(nodeAddress)) //this line will crash if we change db and wallet file naming convention
	if err != nil {
		color.Red("Error: %s\n", err)
		return
	}

	wallet := wallets.GetWallet(from)
	if wallet == nil {
		color.Red("You do not own the address: %s", from)
		return
	}

	// later may be modified to transfer labels in batch
	tx, err := NewUTXOTransaction(wallet, to, serialNumber, salt, &UTXOSet)

	if err != nil {
		color.Red("Error: %s\n", err)
		return
	}

	for _, nodeAddr := range knownNodes {
		sendTx(nodeAddr, tx)
	}

	fmt.Println("Success!")
}

func clientGetHandler(serialNumber, salt string, bc *Blockchain) {

	hash := HashSerialNumber(serialNumber, salt)

	outputs, txIDs := bc.FindSerialNumberHash(hash)

	if len(outputs) == 0 {
		fmt.Printf("No transaction found\n\n")
	}
	
	for index, output := range outputs {
		fmt.Printf("============ Transaction Output ============\n")
		fmt.Printf("txid: %s\n", txIDs[index])
		fmt.Printf("Serial Number Hash: %x\n", output.SerialNumberHash)
		fmt.Printf("Script (PubKey hash of recipient): %x\n", output.PubKeyHash)
		fmt.Printf("\n")
	}
}

func clientPrintHandler(bc *Blockchain) {
	printKnownNodes()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func printKnownNodes() {
	color.Blue("There are %d known nodes now:\n", len(knownNodes))
	color.Blue("%v\n", knownNodes)
}


// ==================================
// ========== Node Network ==========
// ==================================
func startNodeCommunication(nodeAddress string, bc *Blockchain) {
	listener, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error: %s\n", err)
		}
		go handleConnection(conn, bc)
	}
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}

func sendAddr(address string) {
	nodes := addr{knownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)

	sendData(address, request)
}

func sendBlock(addr string, b *Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request)
}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		//remove the unavailable node
		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}
		knownNodes = updatedNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(address, request)
}

func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(address, request)
}

func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}

func sendTx(addr string, tnx *Transaction) {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(addr, request)
}

func sendVersion(addr string, bc *Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = addNewNodes(knownNodes, payload.AddrList)
	printKnownNodes()
	//requestBlocks()
}


func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	//TODO: bc.ValidateBlock().
	if bytes.Compare(bc.tip, block.Hash) == 0 || bc.ValidateBlock(block) == false {
		return
	}

	mutex.Lock()
	bc.AddBlock(block)
	color.Magenta("Added block %x\n", block.Hash)
	mutex.Unlock()

	// server node broadcasts the new block to other nodes
	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress {
				sendInv(node, "block", [][]byte{block.Hash})
			}
		}
	}

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	} else {
		mutex.Lock()
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
		mutex.Unlock()
	}
}

func handleInv(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	color.Magenta("Received inventory with %d %s from %s\n", len(payload.Items), payload.Type, payload.AddrFrom)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		// TODO: download block history from a random known node, instead of the tracker node
		blockHash := payload.Items[0] //block tip hash
		sendGetData(payload.AddrFrom, "block", blockHash)
		
		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func handleGetBlocks(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

// TODO: check we actually have the block or tx
func handleGetData(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		sendTx(payload.AddrFrom, &tx)
		// delete(mempool, txID)
	}
}

func handleTx(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if isMiner {
		txData := payload.Transaction
		tx := DeserializeTransaction(txData)
		mempool[hex.EncodeToString(tx.ID)] = tx

	/*if nodeAddress == knownNodes[0] { //if server node, broadcast tx to all knownNodes
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	}*/
		if len(mempool) >= mining_threshold {
		MineTransactions:
			var txs []*Transaction

			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			mutex.Lock()
			newBlock := bc.MineBlock(txs)

			if newBlock.Height < bc.GetBestHeight() {
				color.Magenta("Lagging behind; trying to catch up the merkle tree")
				mempool = make(map[string]Transaction)
				return
			}
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex()
			mutex.Unlock()

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}

			color.Magenta("New block is mined!")
			sendInv(knownNodes[0], "block", [][]byte{newBlock.Hash})
			/*for _, node := range knownNodes { // broadcast new block to nodes
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}*/

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func handleVersion(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight { //if behind, get new blocks
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight { //if ahead, send version
		sendVersion(payload.AddrFrom, bc)
	}

	sendAddr(payload.AddrFrom)
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}


func handleConnection(conn net.Conn, bc *Blockchain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}

	command := bytesToCommand(request[:commandLength])

	switch command {
	case "addr":
		handleAddr(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	default:
		fmt.Println("Unknown command!")
	}
	conn.Close()
}


// ======================================
// ========== Helper Functions ==========
// ======================================
func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}

// O(n) in practice, since len(newNodes) usually is one.
func addNewNodes(nodes, newNodes []string) []string {
	for _, nodeAddr := range newNodes {
		if !nodeIsKnown(nodeAddr) {
			nodes = append(nodes, nodeAddr)
		}
	}
	return nodes
}




