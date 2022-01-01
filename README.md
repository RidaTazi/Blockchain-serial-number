# Creating a Blockchain to track serial number
### Cas de lutte contre la contrefaçon des médicaments


# Intro:
This blockchain application was forked from https://github.com/Jeiwan/blockchain_go, a simplified implementation of the Bitcoin Core protocol, to suit our specific need of transferring serial numbers.

# Setup: 
- Install go in /usr/local and move the project to /usr/local/src/blockchain_go
```bash
make setup
make build
```

# Scenario:
### 1. Start "server/tracker" node. 
All clients can add and update. This node is server-like only in the sense that it takes new blocks from miners and broadcasts them to client nodes. It is still the fastest miner winning. The concensus mechanism isn't fully decentralized for now.

- ./server.sh localhost 3000 localhost 2000

### 2. Start miner. 
We configured the miner to start mining after receiving two tx. In a new window, type:

- ./miner.sh localhost 2999

### 3. Start client one and create two serial numbers. In a new window, type:

- ./wallet.sh localhost 3001

In the dummy CLI, type:
- add [paste node address from stdout] 3001 0   (3001 is the serial number; 0 is a placeholder for the legacy 'salt' field. We may use this field to add privacy, but let's keep the placeholder now)
- add [paste node address from stdout] 30013001 0

### 4. Start client two, create a serial number, and transfer an existing serial number. In a new window, type:

- ./wallet.sh localhost 3002

In the dummy CLI, type:
- add [paste node address from stdout] 3002 0

Back to client one's window, type:
- send [node one addr] [node two addr] 30013001 0  (transfer '30013001' from node one to node two)

### 5. Start client three, create a serial number, and transfer the '30013001' serial number again. In a new window, type:

- ./wallet.sh localhost 3003

In the dummy CLI, type:
- add [paste node address from stdout] 3003 0
- send [node two addr] [node three addr] 30013001 0  (transfer '30013001' from node two to node three)


### Other things to try in the dummy CLI:
- get 30013001 0  (get [serial number] [salt])
- print