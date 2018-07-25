#!/usr/bin/env bash

# Usage:
#> bin/bot_build
#
# Cross compile (Unix):
#> GOOS=linux GOARCH=amd64 bin/bot_build
#
# Cross compile (OSX):
#> GOOS=darwin GOARCH=amd64 bin/bot_build
#
# Cross compile (Windows):
#> GOOS=windows GOARCH=amd64 bin/bot_build

export GOOS=${GOOS:-`go env GOHOSTOS`}
export GOARCH=${GOARCH:-`go env GOHOSTARCH`}
export GOBIN=`pwd`/build
echo "Compiling 'myst telegram bot' for '$GOOS/$GOARCH'.."

go install myst-bot.go
if [ $? -ne 0 ]; then
    printf "\e[0;31m%s\e[0m\n" "Compile failed!"
    exit 1
fi

exit 0
