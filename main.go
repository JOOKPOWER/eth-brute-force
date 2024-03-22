package main

import (
	"bip44"
	"context"
	"crypto/ecdsa"
	"ethbruteforce/ethbasedclient"
	"ethbruteforce/ierc20"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/tinh98/go-ethereum-hdwallet"
	"math/rand"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"log"
	"math"
	"math/big"
	"os"

	"github.com/aherve/gopool"
	"github.com/ethereum/go-ethereum/common"
)

const (
	POSSIBLE = "0123456789abcdef"
)

var counter = uint64(0)
var maxCheck = uint64(100_000000)
var maxConcurrency = 100

//var client *ethbasedclient.EthBasedClient
//var usdtContract *ierc20.IERC20

const rawURL = "http://202.61.239.89:8545/" //I checking bsc

func main() {
	//flag.Uint64Var(&maxCheck, "maxCheck", 100000, "maximum num address check")
	//flag.IntVar(&maxConcurrency, "maxConcurrency", 150, "maximum num thread")
	fmt.Println("maxCheck", maxCheck)
	fmt.Println("maxConcurrency", maxConcurrency)
	chExit := make(chan os.Signal)

	signal.Notify(chExit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-chExit
		cleanup()
		os.Exit(0)
	}()

	go func() {
		tick := time.NewTicker(10 * time.Second)
		for {
			select {
			// Case statement
			case <-tick.C:
				fmt.Println("num addresses checked :", atomic.LoadUint64(&counter))
			}
		}
	}()

	pool := gopool.NewPool(maxConcurrency)
	for {

		pool.Add(1)
		go func(pool *gopool.GoPool) {
			defer func() {
				if err := recover(); err != nil {
					//fmt.Println("total run ", counter)
					//salvaLog(fmt.Sprintf("total check at %v : %d", time.Now(), counter))
					//chExit <- os.Kill
				}
			}()
			defer pool.Done()
			//fmt.Println("Testado: ", contagem)
			gerar()
		}(pool)
	}
	pool.Wait()
}

var bitSize = 128

func gerar() {
	mnemonic, _ := bip44.NewMnemonic(bitSize)
	seed, err := mnemonic.NewSeed("") // Here you can choose to pass in the specified password or empty string , Different passwords generate different mnemonics
	if err != nil {
		log.Fatal(err)
	}

	wallet, err := hdwallet.NewFromSeed(seed)
	if err != nil {

		log.Fatal(err)
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0") // The last digit is the address of the same mnemonic word id, from 0 Start , The same mnemonic can produce unlimited addresses
	account, err := wallet.Derive(path, false)
	if err != nil {

		log.Fatal(err)
	}

	//TEST
	//privateKey := generateRandomPrivKey()
	//address := generateAddressFromPrivKey(privateKey)
	//address = "0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5"

	address := account.Address.Hex()
	privateKey, _ := wallet.PrivateKeyHex(account)
	//publicKey, _ := wallet.PublicKeyHex(account)

	//fmt.Println("checking ", address, atomic.LoadUint64(&counter))

	client, err := ethbasedclient.NewClient(rawURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	usdtContract, _ := ierc20.NewIERC20(common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"), client.Client)
	balance, err := client.Client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	if balance.Cmp(big.NewInt(0)) != 0 {
		fbalance := new(big.Float)
		fbalance.SetString(balance.String())
		ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
		fmt.Println("address:", address, "Balance ETH:", ethValue)
		strinvalue := fmt.Sprintf("%f", ethValue)
		//salvaLog(mnemonic.Value)
		salvaLog(address)
		salvaLog(privateKey)
		//salvaLog(publicKey)
		salvaLog(strinvalue + " ETH")
		salvaLog("-----------------------------------------------------")
	}

	if balanceFiat, ok := tokenFiat(usdtContract, common.HexToAddress(address)); ok {
		fbalance := new(big.Float)
		fbalance.SetString(balanceFiat.String())
		ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(6)))
		strinvalue := fmt.Sprintf("%f", ethValue)
		//salvaLog(mnemonic.Value)
		salvaLog(address)
		salvaLog(privateKey)
		//salvaLog(publicKey)
		salvaLog(strinvalue + " USDT")
		salvaLog("-----------------------------------------------------")
	}
	atomic.AddUint64(&counter, 1)
	if counter >= maxCheck {
		panic("DONE")
	}
}

func tokenFiat(erc20Contract *ierc20.IERC20, account2 common.Address) (_ *big.Int, ok bool) {
	balance, err := erc20Contract.BalanceOf(&bind.CallOpts{}, account2)
	if err != nil {
		fmt.Println(err)
	}
	if balance.Cmp(big.NewInt(0)) != 0 {
		ok = true
		fbalance := new(big.Float)
		fbalance.SetString(balance.String())
		ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(6)))
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
func generateRandomPrivKey() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var randHex string

	for c := 0; c < 64; c++ {
		n := r.Intn(16)
		randHex += string(POSSIBLE[n])
	}

	return randHex
}

func generateAddressFromPrivKey(hex string) string {
	privateKey, err := crypto.HexToECDSA(hex)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return address
}

func cleanup() {
	fmt.Println("Total addresses:", atomic.LoadUint64(&counter))
	salvaLog(fmt.Sprintf("total addresses check at %v : %d", time.Now(), atomic.LoadUint64(&counter)))
}
