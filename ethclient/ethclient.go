package ethclient

import (
	"context"
	"flag"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mysterium/myst-telegram-bot/keystore"
)

var GethUrl = flag.String("geth.url", "https://ropsten.infura.io/v3/0cf3087cfc4f4c80a349c305aed2d835", "URL value of started geth to connect")

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

func TransferFunds(aa *keystore.FaucetAccount, to *accounts.Account) error {
	client, _, err := LookupBackend()
	if err != nil {
		return err
	}

	nonce, err := client.PendingNonceAt(context.Background(), aa.Account.Address)
	if err != nil {
		return err
	}
	amount := big.NewInt(10000000000000000) // 0.01 ether
	gasLimit := uint64(100000)
	gasPrice := big.NewInt(20000000000) // 20 gwei

	tx := types.NewTransaction(nonce, to.Address, amount, gasLimit, gasPrice, nil)

	signer := types.HomesteadSigner{}
	signature, err := aa.Keystore.SignHash(*aa.Account, signer.Hash(tx).Bytes())
	signedTx, err := tx.WithSignature(signer, signature)

	return client.SendTransaction(context.Background(), signedTx)
}
