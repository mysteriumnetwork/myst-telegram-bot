package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/mysterium/myst-telegram-bot/account"
	"github.com/mysterium/myst-telegram-bot/bot"
)

var cmd = flag.String("cmd", "run", "Command to execute: run, help")
var help = flag.Bool("help", false, "print out help")

func main() {
	flag.Parse()
	err := executeCommand(*cmd)
	if err != nil {
		log.Printf("error occuried: %v\n", err)
		os.Exit(-1)
	}
}

func executeCommand(cmd string) error {
	if *help {
		flag.Usage()
		return nil
	}

	log.Println("Executing with args: " + strings.Join(os.Args[1:], " "))

	switch cmd {
	case "help":
		flag.Usage()
		return nil
	case "run":
		return run()
	default:
		flag.Usage()
		return nil
	}
	return errors.New("unknown command: " + cmd)
}

func run() error {
	faucetAccount, err := account.CreateFaucetAccount()
	if err != nil {
		return err
	}

	bot, err := bot.CreateBot(faucetAccount)
	if err != nil {
		log.Panicln(err)
	}

	log.Println("using account: ", bot.FaucetAccount.Acc.Address.String())
	log.Printf("authorized with bot: %s", bot.Api.Self.UserName)

	err = bot.UpdatesProcessingLoop()
	if err != nil {
		return err
	}
	return nil
}
