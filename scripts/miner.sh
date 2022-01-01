#!/bin/bash

# script Node_IP Node_Port Miner_PubKey
# ./miner.sh localhost 2999 xxxxxxxxx
export NODE_ADDR=$1:$2
export NODE_ID=$2

cd /usr/local/src/blockchain_go
DB_FILE=db/blockchain_${NODE_ID}.db
WALLET_FILE=wallet/wallet_${NODE_ID}.dat

if [ ! -d ./db ]; then
 	mkdir db
fi

if [ ! -d ./wallet ]; then
	mkdir wallet
fi

RUN=bin/bcg

if ! [ -e $DB_FILE ]; then
	addr=$($RUN createwallet)
	addr=${addr#Your new address: } #changing fmt.Printf string in source code will break this line
	echo "Your new address: " $addr
	#copy the genesis block before downloading blocks from server node
	cp db/genesis_block.db $DB_FILE

fi

addr=$($RUN listaddresses)

$RUN startnode -miner $addr
