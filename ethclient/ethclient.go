package ethclient

import (
	"context"
	"flag"
	"fmt"
	"math/big"

	"github.com/MysteriumNetwork/payments/mysttoken/generated"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mysterium/myst-telegram-bot/faucet"
)

var GethUrl = flag.String("geth.url", "https://ropsten.infura.io/v3/0cf3087cfc4f4c80a349c305aed2d835", "URL value of started geth to connect")
var erc20contract = flag.String("erc20.contract", "0x453c11c058f13b36a35e1aee504b20c1a09667de", "Address of ERC20 mintable token")

func LookupBackend() (*ethclient.Client, chan bool, error) {
	ethClient, err := ethclient.Dial(*GethUrl)
	if err != nil {
		return nil, nil, err
	}

	block, err := ethClient.BlockByNumber(context.Background(), nil)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("Latest known block is: ", block.NumberU64())

	progress, err := ethClient.SyncProgress(context.Background())
	if err != nil {
		return nil, nil, err
	}
	completed := make(chan bool)
	if progress != nil {
		fmt.Println("Client is in syncing state - any operations will be delayed until finished")
		// go trackGethProgress(ethClient, progress, completed)
	} else {
		fmt.Println("Geth process fully synced")
		close(completed)
	}

	return ethClient, completed, nil
}

func PrintBalance(account *accounts.Account) error {
	client, _, err := LookupBackend()
	if err != nil {
		return err
	}
	balance, err := client.BalanceAt(context.Background(), account.Address, nil)
	if err != nil {
		return err
	}
	fmt.Println("Your balance is:", balance.String(), "wei")
	return nil
}

func TransferFundsViaPaymentsABI(aa *faucet.FaucetAccount, to *common.Address) error {
	client, _, err := LookupBackend()
	if err != nil {
		return err
	}

	erc20token, err := generated.NewMystTokenTransactor(common.HexToAddress(*erc20contract), client)
	if err != nil {
		return err
	}

	transactor := aa.CreateNewKeystoreTransactor()
	value := big.NewInt(100) // 100 Myst tokens
	signedTx, err := erc20token.Transfer(transactor, *to, value)
	err = client.SendTransaction(context.Background(), signedTx)
	return err
}

func TransferFunds(aa *faucet.FaucetAccount, to *common.Address) error {
	client, _, err := LookupBackend()
	if err != nil {
		return err
	}

	nonce, err := client.PendingNonceAt(context.Background(), aa.Acc.Address)
	if err != nil {
		return err
	}
	value := big.NewInt(0) // 0 ether
	gasLimit := uint64(100000)
	gasPrice := big.NewInt(20000000000) // 20 gwei

	tokenContractAddress := common.HexToAddress("0x453c11c058f13b36a35e1aee504b20c1a09667de")
	paddedToAddress := common.LeftPadBytes(to.Bytes(), 32)
	transferFnSignature := []byte("transfer(address,uint256)")

	// calc transfer fn signature hash
	hash := sha3.NewKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println("MYST token contract transfer method signature hash: ", hexutil.Encode(methodID))

	// set token amount
	amount := new(big.Int)
	amount.SetString("100000000000000000000", 10) // 100 tokens
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	// form token data array
	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedToAddress...)
	data = append(data, paddedAmount...)

	tx := types.NewTransaction(nonce, tokenContractAddress, value, gasLimit, gasPrice, data)

	// sign transaction without private key
	signer := types.HomesteadSigner{}
	signature, err := aa.KS.SignHash(*aa.Acc, signer.Hash(tx).Bytes())
	signedTx, err := tx.WithSignature(signer, signature)

	return client.SendTransaction(context.Background(), signedTx)
}
