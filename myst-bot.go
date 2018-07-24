package main

import (
	"flag"
	"log"

	"github.com/mysterium/myst-telegram-bot/bot"
	"github.com/mysterium/myst-telegram-bot/faucet"
)

var erc20contract = flag.String("erc20.address", "", "Address of ERC20 mintable token")
var amount = flag.Int64("amount", 1000, "Amount of tokens to mint")

func main() {
	faucetAccount, err := faucet.CreateFaucetAccount()
	if err != nil {
		log.Panicln(err)
	}

	bot, err := bot.CreateBot(faucetAccount)
	if err != nil {
		log.Panicln(err)
	}

	//bot.Debug = true
	log.Println("using account: ", bot.FaucetAccount.Acc.Address.String())
	log.Printf("Authorized on account %s", bot.Api.Self.UserName)

	err = bot.UpdatesProcessingLoop()
	if err != nil {
		log.Panicln(err)
	}
}
