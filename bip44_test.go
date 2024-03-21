package main

import (
	"bip44"
	"fmt"
	hdwallet "github.com/tinh98/go-ethereum-hdwallet"
	"log"
	"testing"
)

func TestBip44MetaMask(t *testing.T) {
	bitSize := 128
	mnemonic, _ := bip44.NewMnemonic(bitSize)
	seed, err := mnemonic.NewSeed("password") // Here you can choose to pass in the specified password or empty string , Different passwords generate different mnemonics
	if err != nil {
		log.Fatal(err)
	}

	wallet, err := hdwallet.NewFromSeed(seed)
	if err != nil {

		log.Fatal(err)
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0") // The last digit is the address of the same mnemonic word id, from 0 Start , The same mnemonic can produce unlimited addresses
	account, err := wallet.Derive(path, false)
	if err != nil {

		log.Fatal(err)
	}

	address := account.Address.Hex()
	privateKey, _ := wallet.PrivateKeyHex(account)
	//publicKey, _ := wallet.PublicKeyHex(account)

	fmt.Println(mnemonic.Value)
	fmt.Println(address, privateKey)
}

func TestSaveLog(t *testing.T) {
	salvaLog("ok")
}
