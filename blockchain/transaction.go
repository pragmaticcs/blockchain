package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math/big"

	"github.com/pragmaticcs/blockchain/crypto"
	"github.com/pragmaticcs/blockchain/wallet"
)

type TransactionID = [32]byte

var TransactionZeroID TransactionID

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
	if err != nil {
		return nil, err
	}
	if sum < int64(amount) {
		return nil, wallet.ErrorInsufficientBalance
	}
	for txID, outIDarray := range outs {
		for _, outID := range outIDarray {
			in := TxInput{
				TxID:   txID,
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
		ID:      TransactionZeroID,
		Outputs: outputs,
		Inputs:  inputs,
	}
	txID, err := tx.GetHash()
	if err != nil {
		return nil, err
	}
	tx.ID = txID
	err = b.SignTransaction(tx, sender.PrivateKey)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (t *Transaction) GetHash() ([32]byte, error) {
	tx := *t
	tx.ID = [32]byte{}
	data, err := tx.GetData()
	if err != nil {
		return [32]byte{}, err
	}
	hash := sha256.Sum256(data)
	return hash, nil
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

func NewCoinbase(value int64, to crypto.PubKey) (*Transaction, error) {
	pubKeyHash, err := crypto.HashPublicKey(to)
	if err != nil {
		return nil, err
	}
	txOut := TxOutput{
		Value:      value,
		PubKeyHash: pubKeyHash,
	}
	txIn := TxInput{
		TxID:  [32]byte{}, // Empty ID
		OutID: -1,         // No Reference
		Sig:   nil,
	}
	coinbase := &Transaction{
		ID:      [32]byte{},
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
	return len(t.Inputs) == 1 && t.Inputs[0].TxID == TransactionZeroID && t.Inputs[0].OutID == -1
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

func (tx *Transaction) Copy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput
	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{
			TxID:   in.TxID,
			OutID:  in.OutID,
			Sig:    in.Sig,
			PubKey: in.PubKey,
		})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{
			Value:      out.Value,
			PubKeyHash: out.PubKeyHash,
		})
	}
	return Transaction{
		ID:      tx.ID,
		Inputs:  inputs,
		Outputs: outputs,
	}
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTxs map[TransactionID]Transaction) error {
	if tx.IsCoinbase() {
		return NewTransactionError("coinbase can't be signed")
	}

	for _, in := range tx.Inputs {
		// Check if inputs are valid
		if prevTxs[in.TxID].ID == TransactionZeroID {
			return NewTransactionError("previous transaction doesn't exist")
		}
	}
	txCopy := tx.Copy()

	for id, in := range tx.Inputs {
		prevTx := prevTxs[in.TxID]
		txCopy.Inputs[id].Sig = nil
		txCopy.Inputs[id].PubKey = prevTx.Outputs[in.OutID].PubKeyHash

		txID, err := txCopy.GetHash()
		if err != nil {
			return err
		}

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txID[:])
		if err != nil {
			return err
		}
		sig := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[id].Sig = sig
	}
	return nil
}

func (tx *Transaction) Verify(prevTxs map[TransactionID]Transaction) (bool, error) {
	if tx.IsCoinbase() {
		return true, nil
	}

	for _, in := range tx.Inputs {
		// Check if inputs are valid
		if prevTxs[in.TxID].ID == TransactionZeroID {
			return false, NewTransactionError("previous transaction doesn't exist")
		}
	}
	txCopy := tx.Copy()
	curve := elliptic.P256()

	for id, in := range txCopy.Inputs {
		prevTx := prevTxs[in.TxID]
		txCopy.Inputs[id].Sig = nil
		txCopy.Inputs[id].PubKey = prevTx.Outputs[in.OutID].PubKeyHash
		txID, err := txCopy.GetHash()
		if err != nil {
			return false, err
		}

		var r big.Int
		var s big.Int
		sigLen := len(in.Sig)
		rFromSig := in.Sig[:sigLen/2]
		sFromSig := in.Sig[sigLen/2:]

		r.SetBytes(rFromSig)
		s.SetBytes(sFromSig)

		var x big.Int
		var y big.Int

		keyLen := len(in.PubKey)
		xFromSig := in.PubKey[:keyLen/2]
		yFromSig := in.PubKey[keyLen/2:]

		x.SetBytes(xFromSig)
		y.SetBytes(yFromSig)

		ecdsaPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}

		if !ecdsa.Verify(&ecdsaPubKey, txID[:], &r, &s) {
			return false, nil
		}
	}
	return true, nil
}

type UTXO map[TransactionID][]TxOutput

// UTXO is a map of txID <==> list of unspent outputs

func (b *BlockChain) FindUtxoSet() (UTXO, error) {
	UTXO := make(UTXO)
	spentTXOs := make(map[TransactionID][]int)

	blockchainIterator := b.ToIterator()
	for {
		block := blockchainIterator.Prev()

		if block == nil {
			break
		}

		// For each transaction in a block
		for _, tx := range block.Transactions {
			// Get the txID in string form
			txID := tx.ID

			// OUTPUTS - collect all outputs of the transaction
			// For each output in this transaction
		Outputs:
			for i, txOut := range tx.Outputs {
				// Check if transaction is in the set of [tx <==> spent outputs in tx]
				spentOutsIDs, ok := spentTXOs[txID]
				if ok {
					// If transaction is in the set
					// for each unspent transaction
					for _, spentOutID := range spentOutsIDs {
						// Check if current txOut is already spent
						if spentOutID == i {
							// If it is spent -> continue with the next output
							continue Outputs
						}
					}
				}
				// Check if the output is owned by the owner of pubkey
				UTXO[txID] = append(UTXO[txID], txOut)
			}
			// INPUTS - remove outputs that have already been spent
			// Ignore coinbases
			if !tx.IsCoinbase() {
				// For each input in the transaction's list of inputs
				for _, in := range tx.Inputs {
					// Add the output the input uses to the set of spent outputs
					spentTXOs[in.TxID] = append(spentTXOs[in.TxID], in.OutID)
				}
			}
		}
	}
	return UTXO, nil
}

func (b *BlockChain) FindTransaction(id TransactionID) (*Transaction, error) {
	iter := b.ToIterator()

	for currBlock := iter.Prev(); currBlock != nil; currBlock = iter.Prev() {
		for _, tx := range currBlock.Transactions {
			if bytes.Equal(tx.ID[:], id[:]) {
				return tx, nil
			}
		}
	}
	return nil, NewTransactionError("transaction doesn't exist")
}

func (b *BlockChain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) error {
	prevTxs := make(map[TransactionID]Transaction)

	for _, in := range tx.Inputs {
		prevTx, err := b.FindTransaction(in.TxID)
		if err != nil {
			return err
		}
		prevTxs[prevTx.ID] = *prevTx
	}
	return tx.Sign(privateKey, prevTxs)
}

func (b *BlockChain) VerifyTransaction(tx *Transaction) (bool, error) {
	prevTxs := make(map[TransactionID]Transaction)

	for _, in := range tx.Inputs {
		prevTx, err := b.FindTransaction(in.TxID)
		if err != nil {
			return false, err
		}
		prevTxs[prevTx.ID] = *prevTx
	}
	return tx.Verify(prevTxs)
}

func (b *BlockChain) FindUtxos(pubKeyHash []byte) ([]TxOutput, error) {
	var utxos []TxOutput
	utxoSet, err := b.FindUtxoSet()
	if err != nil {
		return nil, err
	}
	for _, outs := range utxoSet {
		for _, out := range outs {
			if out.IsLockedWithKey(pubKeyHash) {
				utxos = append(utxos, out)
			}
		}
	}
	return utxos, nil
}

func (b *BlockChain) FindSpendableAmount(pubKeyHash []byte) (int, error) {
	// Returns the balance of an account with pubKeyHash
	utxos, err := b.FindUtxos(pubKeyHash)
	if err != nil {
		return -1, err
	}
	var balance int64 = 0
	for _, utxo := range utxos {
		balance += utxo.Value
	}
	return int(balance), nil
}

func (b *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int64, map[TransactionID][]int, error) {
	// Returns a hash map of transaction id / array of indexes of outputs
	utxoSet, err := b.FindUtxoSet()
	spendableOuts := make(map[TransactionID][]int)
	if err != nil {
		return 0, nil, err
	}
	var currSum int64 = 0
Outer:
	for txID, utxos := range utxoSet {
		for id, out := range utxos {
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
