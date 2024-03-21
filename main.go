package main

import (
	"context"
	"fmt"
	"github.com/edunuzzi/go-bip44"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	//"github.com/tyler-smith/go-bip39"
	"log"
	"math"
	"math/big"
	"os"

	"github.com/aherve/gopool"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var contagem = 0

func main() {
	pool := gopool.NewPool(12)
	for {
		pool.Add(1)
		go func(pool *gopool.GoPool) {
			defer pool.Done()
			contagem++
			fmt.Println("Testado:", contagem)
			gerar()
		}(pool)
	}
	pool.Wait()
}

func gerar() {
	bitSize := 256
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
	publicKey, _ := wallet.PublicKeyHex(account)

	client, err := ethclient.Dial("https://bsc-dataseed.binance.org/")
	if err != nil {
		fmt.Println(err)
	}

	account2 := common.HexToAddress(address)
	balance, err := client.BalanceAt(context.Background(), account2, nil)
	if err != nil {
		fmt.Println(err)
	}

	if len(balance.Bits()) == 0 {
		fmt.Println("mnemonic:", mnemonic.Value, "Balance:", balance)

	} else {
		fbalance := new(big.Float)
		fbalance.SetString(balance.String())
		ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
		fmt.Println("mnemonic:", mnemonic.Value, "Balance:", ethValue)
		strinvalue := fmt.Sprintf("%f", ethValue)
		salvaLog(mnemonic.Value)
		salvaLog(address)
		salvaLog(privateKey)
		salvaLog(publicKey)
		salvaLog(strinvalue + " ETH")
		salvaLog("-----------------------------------------------------")

	}
}

func salvaLog(texto string) {
	f, err := os.OpenFile("retorno.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(texto + "\n")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
