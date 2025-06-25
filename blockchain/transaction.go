package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/pragmaticcs/blockchain/crypto"
	"github.com/pragmaticcs/blockchain/wallet"
)

type TransactionID = []byte

type Transaction struct {
	ID      TransactionID
	Inputs  []TxInput
	Outputs []TxOutput
}

func (b *BlockChain) NewTransaction(from, to wallet.Address, amount int, wallets *wallet.Wallets) (*Transaction, error) {
	var inputs []TxInput
	var outputs []TxOutput

	sender, err := wallets.GetWallet(from)
	if err != nil {
		return nil, err
	}
	pubKeyHash, err := crypto.HashPublicKey(sender.PublicKey)
	if err != nil {
		return nil, err
	}
	sum, outs, err := b.FindSpendableOutputs(pubKeyHash, amount)
	fmt.Println(outs)
	if err != nil {
		return nil, err
	}
	if sum < int64(amount) {
		return nil, NewTransactionError(fmt.Sprintf("account [%s] has insufficient funds for transaction", from))
	}
	for txID, outIDarray := range outs {
		txIDbytes, err := hex.DecodeString(txID)
		if err != nil {
			return nil, err
		}
		for _, outID := range outIDarray {
			in := TxInput{
				TxID:   txIDbytes,
				OutID:  outID,
				Sig:    nil,
				PubKey: sender.PublicKey,
			}
			inputs = append(inputs, in)
		}
	}
	outputs = append(outputs, *NewTxOutput(amount, []byte(to))) // Transfer to reciever
	if sum > int64(amount) {
		outputs = append(outputs, *NewTxOutput(int(sum)-amount, []byte(from))) // Transfer change back
	}

	tx := &Transaction{
		ID:      nil,
		Outputs: outputs,
		Inputs:  inputs,
	}
	txID, err := tx.GetHash()
	if err != nil {
		return nil, err
	}
	tx.ID = txID
	return tx, nil
}

func (t *Transaction) GetHash() ([]byte, error) {
	tx := *t
	tx.ID = []byte{}
	data, err := tx.GetData()
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	return hash[:], nil
}

func (t *Transaction) GetData() ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(*t); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type TxOutput struct {
	Value      int64 // Amount of coins
	PubKeyHash []byte
}

func NewTxOutput(value int, address []byte) *TxOutput {
	out := &TxOutput{
		Value:      int64(value),
		PubKeyHash: nil,
	}
	LockOutput(out, address)
	return out
}

type TxInput struct {
	TxID   TransactionID // Points to the ID of the Transaction
	OutID  int           // Points to the index of Tx in the list of Tx's inside the Transaction [ID]
	Sig    crypto.Signature
	PubKey crypto.PubKey
}

func NewCoinbase(value int64, to crypto.PubKey, sig crypto.Signature) (*Transaction, error) {
	pubKeyHash, err := crypto.HashPublicKey(to)
	if err != nil {
		return nil, err
	}
	txOut := TxOutput{
		Value:      value,
		PubKeyHash: pubKeyHash,
	}
	txIn := TxInput{
		TxID:  nil, // Empty ID
		OutID: -1,  // No Reference
		Sig:   sig,
	}
	coinbase := &Transaction{
		ID:      nil,
		Inputs:  []TxInput{txIn},
		Outputs: []TxOutput{txOut},
	}
	id, err := coinbase.GetHash()
	if err != nil {
		return nil, err
	}
	coinbase.ID = id
	return coinbase, nil
}

func (t *Transaction) IsCoinbase() bool {
	// Checks if a transaction is a Coinbase
	return len(t.Inputs) == 1 && t.Inputs[0].TxID == nil && t.Inputs[0].OutID == -1
}

func (out *TxOutput) IsLockedWithKey(pubkeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubkeyHash)
}

func (in *TxInput) UsesKey(pubkeyHash []byte) (bool, error) {
	lockingHash, err := crypto.HashPublicKey(in.PubKey)
	if err != nil {
		return false, err
	}
	return bytes.Equal(lockingHash, pubkeyHash), nil
}

func LockOutput(out *TxOutput, address []byte) {
	pubkeyHash := crypto.DecodeFromBase58(address)
	pubkeyHash = pubkeyHash[1 : len(pubkeyHash)-4]
	out.PubKeyHash = pubkeyHash
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTxs map[string]Transaction) error {
	if tx.IsCoinbase() {
		return NewTransactionError("coinbase can't be signed")
	}

	for _, in := range tx.Inputs {
		// Check if inputs are valid
		if prevTxs[hex.EncodeToString(in.TxID)].ID == nil {
			return NewTransactionError("previous transaction doesn't exist")
		}
	}
	txCopy := *tx

	for id, in := range txCopy.Inputs {
		prevTx := prevTxs[hex.EncodeToString(in.TxID)]
		txCopy.Inputs[id].Sig = nil
		txCopy.Inputs[id].PubKey = prevTx.Outputs[in.OutID].PubKeyHash
		txID, err := txCopy.GetHash()
		if err != nil {
			return err
		}
		txCopy.Inputs[id].TxID = txID
		txCopy.Inputs[id].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			return err
		}
		sig := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[id].Sig = sig
	}
	return nil
}

func (tx *Transaction) Verify(prevTxs map[string]Transaction) (bool, error) {
	if tx.IsCoinbase() {
		return true, nil
	}

	for _, in := range tx.Inputs {
		// Check if inputs are valid
		if prevTxs[hex.EncodeToString(in.TxID)].ID == nil {
			return false, NewTransactionError("previous transaction doesn't exist")
		}
	}
	txCopy := *tx
	curve := elliptic.P256()

	for id, in := range txCopy.Inputs {
		prevTx := prevTxs[hex.EncodeToString(in.TxID)]
		txCopy.Inputs[id].Sig = nil
		txCopy.Inputs[id].PubKey = prevTx.Outputs[in.OutID].PubKeyHash
		txID, err := txCopy.GetHash()
		if err != nil {
			return false, err
		}
		txCopy.Inputs[id].TxID = txID
		txCopy.Inputs[id].PubKey = nil

		var r big.Int
		var s big.Int
		sigLen := len(in.Sig)
		r.SetBytes(in.Sig[:(sigLen / 2)])
		s.SetBytes(in.Sig[(sigLen / 2):])

		var x big.Int
		var y big.Int

		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		ecdsaPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}

		if !ecdsa.Verify(&ecdsaPubKey, txCopy.ID, &r, &s) {
			return false, nil
		}
	}
	return true, nil
}

func (b *BlockChain) FindUnspentTransactions(pubKeyHash []byte) ([]*Transaction, error) {
	var unspentTxs []*Transaction
	// Each transaction has a corresponding array of ids
	// that point to the spent outputs in the transaction's array of outputs
	spentTXOs := make(map[string][]int)

	blockchainIterator := b.ToIterator()
	for {
		block := blockchainIterator.Backward()

		if block == nil {
			break
		}

		// For each transaction in a block
		for _, tx := range block.Transactions {
			// Get the txID in string form
			txID := hex.EncodeToString(tx.ID)

			// OUTPUTS - collect all outputs of the transaction
			// For each output in this transaction
		Outputs:
			for i, txOut := range tx.Outputs {
				// Check if transaction is in the set of [tx <==> spent outputs in tx]
				spentOutsIDs, ok := spentTXOs[txID]
				if ok {
					// If transaction is in the set
					// for each unspent transaction
					for _, spentOut := range spentOutsIDs {
						// Check if index of output in the array of outputs == its own index ???
						if spentOut == i {
							continue Outputs
						}
					}
				}
				// Check if the output is owned by the owner of pubkey
				if txOut.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, tx)
				}
			}
			// INPUTS - remove outputs that have already been spent
			// Ignore coinbases
			if !tx.IsCoinbase() {
				// For each input in the transaction's list of inputs
				for _, in := range tx.Inputs {
					// Check if the input is owned by the owner of pubkey
					usesKey, err := in.UsesKey(pubKeyHash)
					if err != nil {
						return nil, err
					}
					if usesKey {
						inID := hex.EncodeToString(in.TxID)
						// Add the output the input uses to the set of spent outputs
						spentTXOs[inID] = append(spentTXOs[inID], in.OutID)
					}
				}
			}
		}
	}
	return unspentTxs, nil
}

func (b *BlockChain) FindTransaction(id []byte) (*Transaction, error) {
	iter := b.ToIterator()

	for currBlock := iter.Backward(); currBlock != nil; currBlock = iter.Backward() {
		for _, tx := range currBlock.Transactions {
			if bytes.Equal(tx.ID, id) {
				return tx, nil
			}
		}
	}
	return nil, NewTransactionError("transaction doesn't exist")
}

func (b *BlockChain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) error {
	prevTxs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTx, err := b.FindTransaction(in.TxID)
		if err != nil {
			return err
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = *prevTx
	}
	return tx.Sign(privateKey, prevTxs)
}

func (b *BlockChain) VerifyTransaction(tx *Transaction) (bool, error) {
	prevTxs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTx, err := b.FindTransaction(in.TxID)
		if err != nil {
			return false, err
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = *prevTx
	}
	return tx.Verify(prevTxs)
}

func (b *BlockChain) FindUTXO(pubKeyHash []byte) ([]TxOutput, error) {
	var utxos []TxOutput
	unspentTransactions, err := b.FindUnspentTransactions(pubKeyHash)
	if err != nil {
		return nil, err
	}
	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				utxos = append(utxos, out)
			}
		}
	}
	return utxos, nil
}

func (b *BlockChain) FindSpendableAmount(pubKeyHash []byte) (int, error) {
	utxos, err := b.FindUTXO(pubKeyHash)
	if err != nil {
		return -1, err
	}
	var balance int64 = 0
	for _, utxo := range utxos {
		balance += utxo.Value
	}
	return int(balance), nil
}

func (b *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int64, map[string][]int, error) {
	// Returns a hash map of transaction id / array of indexes of outputs
	unspentTxs, err := b.FindUnspentTransactions(pubKeyHash)
	spendableOuts := make(map[string][]int)
	if err != nil {
		return 0, nil, err
	}
	var currSum int64 = 0
Outer:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for id, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && currSum < int64(amount) {
				currSum += out.Value
				spendableOuts[txID] = append(spendableOuts[txID], id)
				if currSum >= int64(amount) {
					break Outer
				}
			}
		}
	}
	return currSum, spendableOuts, nil
}

type TransactionError struct {
	Err string
}

func NewTransactionError(e string) TransactionError {
	return TransactionError{Err: e}
}

func (e TransactionError) Error() string {
	return fmt.Sprintf("Transaction error: %s", e.Err)
}
