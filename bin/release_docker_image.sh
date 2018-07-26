#!/usr/bin/env bash

IMAGE="myst-telegram-bot_alpine"
REPO="mysteriumnetwork"

printf "Publishing $IMAGE image..\n"

docker tag $IMAGE $REPO/myst-telegram-bot:latest
docker tag $IMAGE $REPO/myst-telegram-bot:latest-alpine

docker push $REPO/myst-telegram-bot:latest
docker push $REPO/myst-telegram-bot:latest-alpine
