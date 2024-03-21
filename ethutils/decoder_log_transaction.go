package ethutils

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"strings"
)

type LogTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
}

type LogSwap struct {
	Amount0In  *big.Int
	Amount1In  *big.Int
	Amount0Out *big.Int
	Amount1Out *big.Int
}

func DecodeTransferLog(logs []*types.Log) []LogTransfer {
	var transferEvents []LogTransfer
	var transferEvent LogTransfer

	transferEventHash := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

	for _, vLog := range logs {
		if strings.Compare(vLog.Topics[0].Hex(), transferEventHash.Hex()) == 0 && len(vLog.Topics) >= 4 {
			func() {
				transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
				transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())
				transferEvent.TokenId = vLog.Topics[3].Big()

				transferEvents = append(transferEvents, transferEvent)
			}()
		}
	}

	return transferEvents
}

func DecodeSwapLog(abi *abi.ABI, logs []*types.Log) ([]LogSwap, error) {
	var swapEvents []LogSwap
	swapEvent := LogSwap{}

	swapHash := crypto.Keccak256Hash([]byte("Swap(address,uint256,uint256,uint256,uint256,address)"))

	for _, vLog := range logs {
		if strings.Compare(vLog.Topics[0].Hex(), swapHash.Hex()) == 0 {
			err := abi.UnpackIntoInterface(&swapEvent, "Swap", vLog.Data)
			if err != nil {
				fmt.Println(err)
			}
			swapEvents = append(swapEvents, swapEvent)
		}
	}

	return swapEvents, nil
}
