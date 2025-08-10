# 🔗 Simple BlockChain implementation written in Go

The goal of this hobby project is to learn more about the inner workings of cryptocurrency by building a simplified blockchain from scratch in Go without sacrificing the principles of clean code and modularity.

## Features ✨

As a learning project, this simple blockchain focuses on the following core concepts:
-   **Distributed ledger logic**: Implements blocks, transactions, proof of work, merkle root and other key aspects of a blockchain.
-   **Cryptographical aspects**: Creation of public/private key pairs using ecdsa secp256k1 algorithm, base58 encoding for generating the crypto address, utilizing hashing algorithms like ripemd160, double SHA256.
-   **Blockchain persistance**: Utilizes the [BadgerDB](https://github.com/hypermodeinc/badger) key-value store for efficiently storing and quering both blocks and UTXOs using outpoint flat keys.
-   **Wallet logic**: Enables the user to create and store wallets (private and public keys) on disk using PEM blocks in a single JSON file and calculate the crypto address.  
-   **Peer-to-peer logic**: Implements a TCP transport layer and an RPC messaging protocol similiar to the  
-   **Node logic**: A full node implementation that leverages Go's concurrency capabilities to handle tasks like responding to incoming messages from peers, synchronization, validating incoming blocks, adding new transactions to the mempool for mining.
-   **Cli logic**: A simple cli for starting/stopping the node daemon, sending amount of coins to an address, checking the balance. Implemented using the [Cobra CLI](https://github.com/spf13/cobra) library. Communicates with the node through an IPC server through a unix socket.

Features that should be improved:
-   **P2P Transport Layer**: Currently RPC messages are encoded with gob. A more efficient approach would be to use gRPC or implement a deterministic serialization method for encoding messages.   
-   **NodeSyncronization Process**: Currently uses blocks-first approach for the initial block download. A more sophisticated way to syncronize is the headers-first approach.
-   ****

## Getting started ⚙️

## CLI Usage

## Acknowledgements 📖

A HUGE thanks to these helpful resources without which this hobby project wouldn't be possible:
-   [Building a Blockchain in Golang](https://youtu.be/mYlHT9bB6OE?si=oerme-gf0N2MfX4P) - A YouTube series by Tensor Programming that was very helpful in the genesis of this project (pun intended).
-   [Learn me a bitcoin](https://learnmeabitcoin.com/) - An incredible resource for both beginners and programmers that want to learn the technical side of Bitcoin.
-   [btcd](https://github.com/btcsuite/btcd) - A full node bitcoin implementation written in Go (golang). 
-   [Mastering Blockchain: Unlocking the Power of Cryptocurrencies, Smart Contracts, and Decentralized Applications](https://a.co/d/gQbf4kK) - A book on blockchain and cryptocurrency by Lorne Lantz, Daniel Cawrey responsible for spiking my curiosity about the inner workings of blockchain.
-   [Bitcoin Developer Docs](https://developer.bitcoin.org/reference/intro.html) - Official Bitcoin documentation for beginners.

