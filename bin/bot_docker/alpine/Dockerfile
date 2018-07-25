FROM golang:1.10-alpine AS builder

# Install packages
RUN apk add --update --no-cache bash gcc musl-dev make \
    && rm -rf /var/cache/apk/*

# Compile application
WORKDIR /go/src/github.com/mysterium/myst-telegram-bot
ADD . .
RUN bin/bot_build.sh


FROM alpine:3.8

# Install packages
RUN apk add --update --no-cache ca-certificates bash \
    && rm -rf /var/cache/apk/*

COPY bin/bot_docker/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

COPY --from=builder /go/src/github.com/mysterium/myst-telegram-bot/build/myst-bot /usr/bin/myst-bot

WORKDIR /var/run/myst-bot
