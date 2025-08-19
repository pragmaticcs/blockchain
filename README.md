# 🔗 Simple BlockChain implementation written in Go

The goal of this hobby project is to learn more about the inner workings of cryptocurrency by building a simplified blockchain from scratch in Go without sacrificing the principles of clean code and modularity.

## Features ✨

As a learning project, this simple blockchain focuses on the following core concepts:
-   **Distributed Ledger Logic**: Implementation of core blockchain concepts such as blocks, transactions, Proof-of-Work, and Merkle roots.
-   **Cryptography**: Generation of public/private key pairs using the `ECDSA (secp256k1)` curve, creation of human-readable addresses via `Base58` encoding, use of hashing algorithms including `SHA-256` (double-hashed) and `RIPEMD-160`.
-   **Blockchain Persistance**: Utilizes the [BadgerDB](https://github.com/hypermodeinc/badger) key-value store for efficiently storing and quering blocks and the UTXO set.
-   **Wallet Management**: Creation and on-disk storage of wallets. Keys are PEM-encoded and stored within a user-friendly JSON file.
-   **Peer-to-Peer Network**: A simple TCP-based transport layer with an RPC protocol (using Go's `gob` for serialization) for peer communication.  
-   **Full Node Implementation**: A full node implementation that leverages Go's concurrency capabilities to handle peer communication, blockchain synchronization, transaction validation, and mempool management for mining.
-   **Cli Logic**: A simple CLI built with [Cobra](https://github.com/spf13/cobra) to manage the node (start/stop), create wallets, check balances, and send transactions. The CLI communicates with the node daemon via an IPC server over a Unix socket.
-   **Docker Support**: Includes a `Dockerfile` for building and running the application in a containerized environment and a `docker-compose.yml` for starting up a blockchain network consisting of 3 nodes. 

## Roadmap for Future Improvements

While the current implementation covers core functionality, the following issues need to be addressed in the future:
-   **P2P Transport Layer**: Currently RPC messages are encoded with gob. A more efficient approach would be to use gRPC or implement a deterministic serialization method for encoding messages.   
-   **NodeSyncronization Process**: Implement a "headers-first" synchronization approach for a more efficient initial block download compared to the current "blocks-first" method.
-   **Command-line Interface**: The cli probably has bugs and issues that need proper testing and further development, as it wasn't my main focus while writing this project.

## Getting started ⚙️

### Prerequisites

-   **Go:** Version 1.20 or higher.
-   **Make:** A `Makefile` is included for simplified build commands.
-   **OS:** Developed and tested on Linux. Docker is recommended for cross-platform compatibility.

### Building from source

1.  Clone the repo:
    ```sh
    foo@bar:~$ https://github.com/pragmaticcs/blockchain.git
    foo@bar:~$ cd blockchain
    ```

2.  Build using make (this will create the executable in the projects bin directory):
    ```sh
    foo@bar:~/blockchain$ make build
    ```

## CLI Usage

The CLI is the primary way to interact with your blockchain node and wallet.

### 1. Creating a Wallet

Before running a node, you must create a wallet. This command creates a wallet named `[name]` in the specified `[directory]`.

```sh
./bin/blockchain wallet --dir [directory] --name [name] new
```

*Example: Creates a wallet and saves it to `data/satoshi/satoshi_wallet.json`*
```sh
./bin/blockchain wallet --dir data/satoshi --name satoshi new
```

### 2. Getting a Wallet Address

To receive funds, you need your wallet's public address.

```sh
./bin/blockchain wallet --dir [directory] --name [name] address
```

*Example:*
```sh
./bin/blockchain wallet --dir data/satoshi --name satoshi address
# Output:
# Wallet address: 1BDK9VHwozyF3CpWdpqkrxtwznJovWWyhb
```

### 3. Running a Node

A full node requires a data directory to store blockchain data and logs. Node behavior is configured using a `config.toml` file placed in the root of this data directory.

*Example `config.toml`:*
```toml
[node]

[node.wallet]
dir = './'       # Specifies the directory where the wallet file is located
name = 'satoshi'  # The name of the wallet to use for mining rewards

[node.server]
port = ":8080"
peers = [
    ":8081",
    ":8082"
]
```

Start the node:
```sh
./bin/blockchain node --dataDir [data-directory] start
```
*Node logs will be written to `gochain.log` within the data directory.*

Stop the node:
```sh
./bin/blockchain node --dataDir [data-directory] stop
```

### 4. Sending Coins

Once your node is running, you can send coins to another address.

```sh
./bin/blockchain node --dataDir [data-directory] send --amount [amount] --fee [fee] --address [recipient-address]
```

## Acknowledgements 📖

A HUGE thanks to these helpful resources without which this hobby project wouldn't be possible:
-   [Building a Blockchain in Golang](https://youtu.be/mYlHT9bB6OE?si=oerme-gf0N2MfX4P) - A YouTube series by Tensor Programming that was very helpful in the genesis of this project (pun intended).
-   [Learn me a bitcoin](https://learnmeabitcoin.com/) -  An incredible resource for learning the technical details of Bitcoin.
-   [btcd](https://github.com/btcsuite/btcd) - A full node bitcoin implementation written in Go (golang), which served as a valuable reference.
-   [Mastering Blockchain: Unlocking the Power of Cryptocurrencies, Smart Contracts, and Decentralized Applications](https://a.co/d/gQbf4kK) - A book by Lorne Lantz and Daniel Cawrey that sparked my curiosity about the inner workings of blockchain.
-   [Bitcoin Developer Docs](https://developer.bitcoin.org/reference/intro.html) - Official Bitcoin documentation.

