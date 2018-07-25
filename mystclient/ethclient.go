package mystclient

import (
	"context"
	"flag"
	"math/big"

	"log"

	"errors"

	"github.com/MysteriumNetwork/payments/mysttoken/generated"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mysterium/myst-telegram-bot/account"
)

var GethUrl = flag.String("geth.url", "https://ropsten.infura.io/v3/0cf3087cfc4f4c80a349c305aed2d835", "URL value of started geth to connect")
var erc20contract = flag.String("erc20.contract", "0x453c11c058f13b36a35e1aee504b20c1a09667de", "Address of ERC20 MYST token contract")
var amount = flag.Int64("amount", 100, "Amount of MYST tokens to transfer")
var maxAmount = flag.Int64("amount.max", 1000, "Maximum target amount of MYST tokens that can be transferred")

var ErrTotalTokensExhausted = errors.New("max tokens transferred to given address reached")
var ErrTooFrequentRequest = errors.New("too frequent duplicate request")

type MystClient struct {
	api               *ethclient.Client
	addressReqByBlock map[string]uint64
}

func Create() (*MystClient, error) {
	ethClient, err := ethclient.Dial(*GethUrl)
	if err != nil {
		return nil, err
	}

	return &MystClient{
		api:               ethClient,
		addressReqByBlock: make(map[string]uint64),
	}, nil
}

func (client *MystClient) PrintBalance(account *accounts.Account) error {
	balance, err := client.api.BalanceAt(context.Background(), account.Address, nil)
	if err != nil {
		return err
	}
	log.Println("MYST faucet balance is:", balance.String(), "wei")
	return nil
}

func (client *MystClient) TransferFundsViaPaymentsABI(aa *account.FaucetAccount, to *common.Address) error {

	if err := client.IsEligibleForTransfer(aa, to); err != nil {
		return err
	}

	log.Printf("sending %d MYST tokens to: %s", *amount, to.String())

	erc20token, err := generated.NewMystTokenTransactor(common.HexToAddress(*erc20contract), client.api)
	if err != nil {
		return err
	}

	transactor := aa.CreateNewKeystoreTransactor()
	value := big.NewInt(*amount)
	_, err = erc20token.Transfer(transactor, *to, value)
	return err
}

func (client *MystClient) IsEligibleForTransfer(aa *account.FaucetAccount, to *common.Address) error {
	err := client.throttleRequests(aa, to)
	if err != nil {
		return err
	}

	mystTokenFilterer, err := generated.NewMystTokenFilterer(common.HexToAddress(*erc20contract), client.api)
	if err != nil {
		return err
	}

	fromAddresses := []common.Address{aa.Acc.Address}
	toAddresses := []common.Address{*to}

	filterer := aa.CreateNewFilterer()
	logIterator, err := mystTokenFilterer.FilterTransfer(filterer, fromAddresses, toAddresses)
	if err != nil {
		return err
	}

	var totalValue big.Int

	for {
		next := logIterator.Next()
		if !next {
			if logIterator.Error() != nil {
				log.Println(logIterator.Error())
			}
			break
		}
		event := logIterator.Event
		totalValue.Add(&totalValue, event.Value)
	}

	log.Println("total tokens already transferred: ", totalValue.Int64())

	if totalValue.Int64() > *maxAmount {
		return ErrTotalTokensExhausted
	}

	return nil
}

func (client *MystClient) throttleRequests(faucetAccount *account.FaucetAccount, address *common.Address) error {
	lastBlock, err := client.api.BlockByNumber(context.Background(), nil)
	if err != nil {
		return err
	}

	reqBlock, present := client.addressReqByBlock[address.String()]
	if !present {
		client.addressReqByBlock[address.String()] = lastBlock.NumberU64()
		return nil
	}

	log.Println("Latest block is: ", lastBlock.NumberU64(), " last known request was on: ", reqBlock)
	// we wait at least 8 blocks for another request to pass
	if reqBlock > lastBlock.NumberU64()-8 {
		return ErrTooFrequentRequest
	}

	client.addressReqByBlock[address.String()] = lastBlock.NumberU64()
	return nil
}
