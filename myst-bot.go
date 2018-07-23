package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"flag"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/mysterium/myst-telegram-bot/keystore"
	"fmt"
	"github.com/mysterium/myst-telegram-bot/ethclient"
	"context"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

var erc20contract = flag.String("erc20.address", "", "Address of ERC20 mintable token")
var amount = flag.Int64("amount", 1000, "Amount of tokens to mint")


func main() {
	aa := &keystore.ActiveAccount{}
	account, err := aa.GetAccount()
	if err != nil {
		log.Panic(err)
	}
	log.Println("using account: ", account.Address.String())

	bot, err := tgbotapi.NewBotAPI("***REMOVED***")
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s %s (%s-%s)] %s", update.Message.From.FirstName, update.Message.From.LastName,
			update.Message.From.UserName, update.Message.From.LanguageCode, update.Message.Text)

		toAccount, err := getEtherAccount(update.Message.Text)
		if err != nil {
			sendBotMessage(bot, update, err.Error())
			log.Println(err)
			continue
		}

		err = printBalance(account)
		if err != nil {
			log.Println(err)
		}

		log.Println("sending 0.01 eth to: ", toAccount.Address.String())
	    err = transferFunds(aa, toAccount)
		if err != nil {
			log.Println(err)
		}
	}
}

func sendBotMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, message string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
	msg.ReplyToMessageID = update.Message.MessageID

	bot.Send(msg)
}

func getEtherAccount(address string) (account *accounts.Account, err error) {
	//TODO: parse real
	address = "0x0a6d6733cf17311499184ab509e4590a09952ba4"
	return addressToAccount(address), nil
}

type etheriumFacetResult struct {
	Paydate int
	Address string
	Amount int
	Message string
	Snapshot int64
	Duration int64
}

func getEthers(address string) (string, error) {
	resp, err := http.Get("http://faucet.ropsten.be:3001/donate/"+address)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// TODO: validate response
	if err != nil {
		return "", err
	}

	var facetResult etheriumFacetResult
	err = json.Unmarshal(body, &facetResult)
	if err != nil {
		log.Println("failed to decode facet response: ", err)
		return "", err
	}

	spew.Dump(facetResult)
	return facetResult.Message, nil
}

func printBalance(account *accounts.Account) error {
	client, _, err := ethclient.LookupBackend()
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

func transferFunds(aa *keystore.ActiveAccount, to *accounts.Account) error {
	client, _, err := ethclient.LookupBackend()
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

func addressToAccount(address string) *accounts.Account {
	return &accounts.Account{
		Address: common.HexToAddress(address),
	}
}
