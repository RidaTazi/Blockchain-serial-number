#!/bin/bash

cd /usr/local/src/blockchain_go
rm -rf db/blockchain_*
rm -rf wallet/wallet_*
echo "All data cleared except genesis_block.db"
