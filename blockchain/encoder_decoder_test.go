package blockchain

// func TestGobEncoderDecoder(t *testing.T) {
// 	handle := func(err error) {
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	}
// 	wallets := &wallet.Wallets{
// 		Wallets:      map[wallet.Address]*wallet.Wallet{},
// 		SaveFilePath: "wallets.json",
// 	}
// 	bob, err := wallets.AddWallet()
// 	handle(err)
// 	bobsWallet, err := wallets.GetWallet(bob)
// 	handle(err)
// 	difficulty := 16
// 	genesisConf := GenesisConfig{Amount: 50, To: bobsWallet.PublicKey, Sig: nil}
// 	bch, err := NewBlockChain(difficulty, genesisConf, wallets)
// 	handle(err)

// 	coinbase, err := NewCoinbase(1000000, []byte("my public key"))
// 	handle(err)
// 	block := bch.AddBlock([]*Transaction{coinbase})

// 	encoderDecoder := GobEncoderDecoder{}
// 	encoding, err := encoderDecoder.EncodeBlock(block)
// 	handle(err)

// 	decoding, err := encoderDecoder.DecodeBlock(encoding)
// 	handle(err)
// 	assert.EqualValues(t, block, decoding)
// }
