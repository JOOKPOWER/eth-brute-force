package ethbasedclient

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"ethbruteforce/errorsutil"
	"ethbruteforce/ierc20"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"time"
)

type EthBasedClient struct {
	Client         *ethclient.Client
	PrivateKey     *ecdsa.PrivateKey
	PublicKeyECDSA *ecdsa.PublicKey
	Address        common.Address
	ChainID        *big.Int
	Transactor     *bind.TransactOpts
	Nonce          *big.Int
}

func NewClient(rawurl string) (*EthBasedClient, error) {
	client, err := ethclient.Dial(rawurl)
	if err != nil {
		return nil, err
	}
	// get chain id
	chainID, chainIDErr := client.ChainID(context.Background())
	if chainIDErr != nil {
		return nil, chainIDErr
	}

	ethBasedClientTemp := &EthBasedClient{
		Client:  client,
		ChainID: chainID,
	}

	return ethBasedClientTemp, nil
}

func New(rawurl string, privateKey *ecdsa.PrivateKey) (*EthBasedClient, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}

	// generate public key and address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	// generate address from public key
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// get chain id
	chainID, chainIDErr := client.ChainID(ctx)
	if chainIDErr != nil {
		return nil, chainIDErr
	}

	// generate transactor for transactions management
	transactor, transactOptsErr := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if transactOptsErr != nil {
		return nil, transactOptsErr
	}

	ethBasedClientTemp := &EthBasedClient{
		Client:         client,
		PrivateKey:     privateKey,
		PublicKeyECDSA: publicKeyECDSA,
		Address:        address,
		ChainID:        chainID,
		Transactor:     transactor,
	}

	return ethBasedClientTemp, nil
}

func (ethBasedClient *EthBasedClient) ConfigureTransactor(value *big.Int, gasPrice *big.Int, gasLimit uint64) {

	if value != nil && value.String() != "-1" {
		ethBasedClient.Transactor.Value = value
	}

	ethBasedClient.Transactor.GasPrice = gasPrice
	ethBasedClient.Transactor.GasLimit = gasLimit
	ethBasedClient.Transactor.Nonce = ethBasedClient.PendingNonce()
	ethBasedClient.Transactor.Context = context.Background()
}

func (ethBasedClient *EthBasedClient) Balance() *big.Int {
	// get current balance
	balance, balanceErr := ethBasedClient.Client.BalanceAt(context.Background(), ethBasedClient.Address, nil)
	errorsutil.HandleError(balanceErr)
	return balance
}

func (ethBasedClient *EthBasedClient) PendingNonce() *big.Int {
	// calculate next nonce
	nonce, nonceErr := ethBasedClient.Client.PendingNonceAt(context.Background(), ethBasedClient.Address)
	errorsutil.HandleError(nonceErr)
	return big.NewInt(int64(nonce))
}

func (ethBasedClient *EthBasedClient) PendingNonceUint64() uint64 {
	// calculate next nonce
	nonce, nonceErr := ethBasedClient.Client.PendingNonceAt(context.Background(), ethBasedClient.Address)
	errorsutil.HandleError(nonceErr)
	return nonce
}

func (ethBasedClient *EthBasedClient) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	// get current balance
	balance, balanceErr := ethBasedClient.Client.BalanceAt(context.Background(), ethBasedClient.Address, nil)
	if balanceErr != nil {
		return nil, balanceErr
	}
	if balance.Cmp(amount) == -1 {
		return nil, core.ErrInsufficientFunds
	}

	nonce, err := ethBasedClient.Client.PendingNonceAt(context.Background(), ethBasedClient.Address)
	if err != nil {
		return nil, err
	}
	gasLimit := uint64(21000)
	gasPrice, err := ethBasedClient.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	var data []byte
	legacyTx := &types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &to,
		Data:     data,
		Value:    amount,
	}
	tx := types.NewTx(legacyTx)

	chainID, err := ethBasedClient.Client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), ethBasedClient.PrivateKey)
	if err != nil {
		return nil, err
	}

	err = ethBasedClient.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func (ethBasedClient *EthBasedClient) TransferERC20(erc20Contract *ierc20.IERC20, to common.Address, value *big.Int) (*types.Transaction, error) {
	return erc20Contract.Transfer(ethBasedClient.Transactor, to, value)
}

const (
	APPROVE_AMOUNT string = "1000000000000000000000000000"
)

func (ethBasedClient *EthBasedClient) CheckAllowance(erc20Contract *ierc20.IERC20, spender common.Address, amount *big.Int) error {
	biInt, err := erc20Contract.Allowance(&bind.CallOpts{}, ethBasedClient.Address, spender)
	if err != nil {
		return err
	}

	if biInt.Cmp(amount) == -1 {
		//get all balance
		bal, err := ethBasedClient.BalanceOfIERC20(erc20Contract)
		if err != nil {
			return err
		}

		if len(bal.Bits()) == 0 {
			return core.ErrInsufficientFunds
		}

		amountApprove, ok := new(big.Int).SetString(APPROVE_AMOUNT, 10)
		if !ok {
			return errors.New("wrong approve amount")
		}

		err = ethBasedClient.Approve(erc20Contract, spender, amountApprove)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (ethBasedClient *EthBasedClient) Approve(erc20Contract *ierc20.IERC20, spender common.Address, amountApprove *big.Int) error {

	approveTx, ApproveErr := erc20Contract.Approve(
		ethBasedClient.Transactor,
		spender,
		amountApprove)
	if ApproveErr != nil {
		return ApproveErr
	}
	txHash := approveTx.Hash().Hex()
	fmt.Println(txHash)
	time.Sleep(3 * time.Second) //wait a minute =)))

	return nil
}

//get balance ierc20 token
func (ethBasedClient *EthBasedClient) BalanceOfIERC20(erc20Contract *ierc20.IERC20) (*big.Int, error) {

	bal, err := erc20Contract.BalanceOf(&bind.CallOpts{}, ethBasedClient.Address)
	if err != nil {
		return nil, err
	}
	return bal, nil
}
