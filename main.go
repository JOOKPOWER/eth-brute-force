package main

import (
	"context"
	"ethbruteforce/ethbasedclient"
	"ethbruteforce/ierc20"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	hdwallet "github.com/tinh98/go-ethereum-hdwallet"
	//"github.com/tyler-smith/go-bip39"
	"bip44"
	"log"
	"math"
	"math/big"
	"os"

	"github.com/aherve/gopool"
	"github.com/ethereum/go-ethereum/common"
)

var contagem = 0

var client *ethbasedclient.EthBasedClient
var usdtContract *ierc20.IERC20

const rawURL = "https://bsc-dataseed1.binance.org/" //I checking bsc

func main() {
	defer fmt.Println("total run ", contagem)
	var err error
	client, err = ethbasedclient.NewClient(rawURL)
	if err != nil {
		panic(err)
	}

	usdtContract, _ = ierc20.NewIERC20(common.HexToAddress("0x55d398326f99059ff775485246999027b3197955"), client.Client)

	pool := gopool.NewPool(12)
	for {
		pool.Add(1)
		go func(pool *gopool.GoPool) {
			defer pool.Done()
			contagem++
			//fmt.Println("Testado: ", contagem)
			gerar()
		}(pool)
	}
	pool.Wait()
}

func gerar() {
	bitSize := 128
	mnemonic, _ := bip44.NewMnemonic(bitSize)
	seed, err := mnemonic.NewSeed("") // Here you can choose to pass in the specified password or empty string , Different passwords generate different mnemonics
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

	//TEST
	//account.Address = common.HexToAddress("0x1c995af606047c4dfA07a1c5E120e57296b04863")

	address := account.Address.Hex()
	privateKey, _ := wallet.PrivateKeyHex(account)
	publicKey, _ := wallet.PublicKeyHex(account)

	//fmt.Println("checking ", address)

	balance, err := client.Client.BalanceAt(context.Background(), account.Address, nil)
	if err != nil {
		fmt.Println(err)
	}

	if len(balance.Bits()) == 0 {

	} else {
		fbalance := new(big.Float)
		fbalance.SetString(balance.String())
		ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
		fmt.Println("address:", address, "Balance ETH:", ethValue)
		strinvalue := fmt.Sprintf("%f", ethValue)
		salvaLog(mnemonic.Value)
		salvaLog(address)
		salvaLog(privateKey)
		salvaLog(publicKey)
		salvaLog(strinvalue + " BNB")
		salvaLog("-----------------------------------------------------")
	}

	if balanceFiat, ok := tokenFiat(usdtContract, account.Address); ok {
		fbalance := new(big.Float)
		fbalance.SetString(balanceFiat.String())
		ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
		strinvalue := fmt.Sprintf("%f", ethValue)
		salvaLog(mnemonic.Value)
		salvaLog(address)
		salvaLog(privateKey)
		salvaLog(publicKey)
		salvaLog(strinvalue + " USDT")
		salvaLog("-----------------------------------------------------")
	}
}

func tokenFiat(erc20Contract *ierc20.IERC20, account2 common.Address) (*big.Int, bool) {
	var ok = false
	balance, err := erc20Contract.BalanceOf(&bind.CallOpts{}, account2)
	if err != nil {
		fmt.Println(err)
	}
	if len(balance.Bits()) == 0 {

	} else {
		ok = true
		fbalance := new(big.Float)
		fbalance.SetString(balance.String())
		ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
		fmt.Println("address: ", account2, "Balance: USDT: ", ethValue)
	}
	return balance, ok
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
