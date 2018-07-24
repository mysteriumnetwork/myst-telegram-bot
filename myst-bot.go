package main

import (
	"log"

	"github.com/mysterium/myst-telegram-bot/account"
	"github.com/mysterium/myst-telegram-bot/bot"
)

func main() {
	faucetAccount, err := account.CreateFaucetAccount()
	if err != nil {
		log.Panicln(err)
	}

	bot, err := bot.CreateBot(faucetAccount)
	if err != nil {
		log.Panicln(err)
	}

	log.Println("using account: ", bot.FaucetAccount.Acc.Address.String())
	log.Printf("authorized with bot: %s", bot.Api.Self.UserName)

	err = bot.UpdatesProcessingLoop()
	if err != nil {
		log.Panicln(err)
	}
}
