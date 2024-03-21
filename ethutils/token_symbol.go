package ethutils

import "strings"

var _tokenSymbolLowerCase = map[string]string{
	"0x55d398326f99059ff775485246999027b3197955": "usdt",
	"0xb8c77482e45f1f44de1745f52c74426c631bdd52": "bnb",
	"0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c": "wbnb",
}

func GetSymbol(tokenAddress string) string {
	symbol, exists := _tokenSymbolLowerCase[strings.ToLower(tokenAddress)]
	if !exists {
		return tokenAddress
	}
	return symbol
}
