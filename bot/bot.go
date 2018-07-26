package bot

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mysterium/myst-telegram-bot/account"
	"github.com/mysterium/myst-telegram-bot/mystclient"
	"gopkg.in/telegram-bot-api.v4"
)

var ErrEtherAddressInvalid = errors.New("invalid ethereum address supplied")
var ErrCommandIncomplete = errors.New("command incomplete")
var ErrBotTokenMissing = errors.New("telegram bot token missing")
var ErrCommandInvalid = errors.New("invalid command, available commands: \n /send 0x_your_ethereum_address - sends some myst tokens to given ropsten testnet account")
var ErrRequestToFast = errors.New("you are sending commands too fast - try later")

var botToken = flag.String("bot.token", "", "telegram bot auth token")

type Bot struct {
	Api               *tgbotapi.BotAPI
	FaucetAccount     *account.FaucetAccount
	UsersRequestTimes map[string]int64
}

func CreateBot(fa *account.FaucetAccount) (*Bot, error) {
	if *botToken == "" {
		return nil, ErrBotTokenMissing
	}

	Api, err := tgbotapi.NewBotAPI(*botToken)
	if err != nil {
		return nil, err
	}

	return &Bot{Api, fa, make(map[string]int64)}, nil
}

func (bot *Bot) UpdatesProcessingLoop() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.Api.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	mystClient, err := mystclient.Create()
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s %s (%s-%s)] %s", update.Message.From.FirstName, update.Message.From.LastName,
			update.Message.From.UserName, update.Message.From.LanguageCode, update.Message.Text)

		err := bot.throttleRequests(update)
		if err != nil {
			bot.sendBotMessage(update, err.Error())
			log.Println(err)
			continue
		}

		toAddress, err := getEtherAddress(update.Message.Text)
		if err != nil {
			bot.sendBotMessage(update, err.Error())
			log.Println(err)
			continue
		}

		err = mystClient.PrintBalance(bot.FaucetAccount.Acc)
		if err != nil {
			log.Println(err)
		}

		err = mystClient.TransferFundsViaPaymentsABI(bot.FaucetAccount, toAddress)
		if err != nil {
			bot.sendBotMessage(update, err.Error())
			log.Println(err)
			continue
		}

		msg := fmt.Sprintf("MYST tokens transfer initiated. Check https://ropsten.etherscan.io/address/%s in a few seconds.", toAddress.String())
		log.Printf("sending command reply: %s", msg)
		bot.sendBotMessage(update, msg)
	}

	return nil
}

func (bot *Bot) throttleRequests(update tgbotapi.Update) error {
	log.Println("user count: ", len(bot.UsersRequestTimes))

	var user string
	if update.Message.From.UserName != "" {
		user = update.Message.From.UserName
	} else {
		user = fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName)
	}

	lastReqTime, present := bot.UsersRequestTimes[user]

	if !present {
		bot.UsersRequestTimes[user] = time.Now().UnixNano()
		return nil
	}

	// we wait at least 10 seconds for next request
	if (time.Now().UnixNano() - lastReqTime) < int64(math.Pow10(10)) {
		return ErrRequestToFast
	}

	bot.UsersRequestTimes[user] = time.Now().UnixNano()
	return nil
}

func (bot *Bot) sendBotMessage(update tgbotapi.Update, message string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
	msg.ReplyToMessageID = update.Message.MessageID

	bot.Api.Send(msg)
}

func getEtherAddress(botText string) (*common.Address, error) {
	botText = strings.TrimSpace(botText)
	command := strings.Fields(botText)

	if len(command) < 2 {
		return nil, ErrCommandInvalid
	}

	address := strings.TrimSpace(command[1])
	log.Println("address: ", address)

	switch command[0] {
	case "/send":
		if !account.IsAddressValid(command[1]) {
			return nil, ErrEtherAddressInvalid
		}
	default:
		return nil, ErrCommandInvalid
	}

	etherAddress := common.HexToAddress(address)
	return &etherAddress, nil
}
