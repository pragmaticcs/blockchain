package blockchain

type TxOutputs struct {
	Outputs []TxOutput
}
type UTXOs map[TransactionID]TxOutputs

type UTXOset struct {
	Blockchain *Blockchain
	Store      Store
}
