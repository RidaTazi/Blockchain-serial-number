#!/bin/bash

# script Node_IP Node_Port API_IP API_Port
# ./server.sh localhost 3000 localhost 2000
export NODE_ADDR=$1:$2
export NODE_ID=$2
export API_ADDR=$3:$4

cd /usr/local/src/blockchain_go
DB_FILE=db/blockchain_${NODE_ID}.db
WALLET_FILE=wallet/wallet_${NODE_ID}.dat

if [ ! -d ./db ]; then
 	mkdir db
fi

if [ ! -d ./wallet ]; then
	mkdir wallet
fi

RUN=./bin/bcg

if ! [ -e $DB_FILE ]; then
	addr=$($RUN createwallet)
	addr=${addr#Your new address: } #changing fmt.Printf string in source code will break this line
	$RUN createblockchain -address $addr
	cp $DB_FILE db/genesis_block.db
	echo "Created server node"
fi

echo "Server Address: " $($RUN listaddresses)

$RUN startnode
