package ethutils

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"time"
)

func CheckTransaction(client *ethclient.Client, txHash common.Hash) (txReceipt *types.Receipt, err error) {
	for {
		time.Sleep(3 * time.Second)
		fmt.Printf("Waiting receipt of transaction %s\n", txHash.Hex())
		if !IsTransactionPending(client, context.Background(), txHash) {
			txReceipt, err = client.TransactionReceipt(context.Background(), txHash)
			if err != nil {
				return
			}
			fmt.Printf("Got transaction receipt status: %v \n", txReceipt.Status)
			break
		}
	}
	return
}

func IsTransactionPending(client *ethclient.Client, context context.Context, hash common.Hash) bool {
	_, pending, err := client.TransactionByHash(context, hash)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return pending
}
