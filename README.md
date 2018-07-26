# Mysterium network testnet faucet telegram bot

This is a bot / faucet for airdropping MYST tokens for ethereum testnet (ropsten) accounts.

Requires go 1.9.2 or later and docker (optional)

First you need to create telegram bot and get its authorization token. To make it happen, in telegram send a message to @BotFather bot.
From there you can type '/help' and proceed with creating your own bot.

## running standalone
```
go run myst-bot.go --bot.token="auth_token_after_bot_created" --ether.address="0x_your_bots_ethereum_address" --keystore.passphrase="ethereum_accout_pass"
```

## with docker

### Building
```
docker-compose build
```
### Running
```
docker run  -v /path_to_ethereum_testnet_keystore/:/var/run/myst-bot/testnet -d --name myst-telegram-bot myst-telegram-bot_alpine:latest --ether.address="0x_your_bots_ethereum_address" --keystore.passphrase="ethereum_accout_pass"
```

### Test bot status
```
docker logs myst-telegram-bot
```

you should see something like this:
```
Executing: run
2018/07/26 07:09:55 Faucet newAccount:  false
2018/07/26 07:09:55 Trying to use account:  0xCf16489612B1D8407Fd66960eCB21941718CD8FD
2018/07/26 07:09:56 using account:  0xCf16489612B1D8407Fd66960eCB21941718CD8FD
2018/07/26 07:09:56 authorized with bot: myst_testnet_faucet_bot
```

Now you can interact with your bot sending message to `@your_bot_name` and typing `/send 0x_account_to_send_tokens_to` command.

For example, to talk with Mysterium Network bot you can send message to `@myst_testnet_faucet_bot`
 and ask to send you some testnet MYST tokens with `/send 0x_your_etherium_testnet_account`
