FROM python:3.13.0-alpine3.20 AS scraper

WORKDIR /app/sushi-roulette/scraper
COPY ./scraper/ .

RUN python3 -m pip install -r requirements.txt
RUN mkdir ../json
RUN python3 hama-sushi.py
RUN python3 uobei.py
RUN python3 sushiro.py

FROM golang:1.23.3-alpine3.20 AS builder

WORKDIR /app/sushi-roulette/bot
COPY ./bot/ .

RUN go mod download
RUN go build bot.go

FROM alpine:3.13 AS release

LABEL maintainer="oka4shi"
WORKDIR /app/sushi-roulette

COPY --from=scraper /app/sushi-roulette/json/* ./dist/
COPY --from=builder /app/sushi-roulette/bot/bot ./bot/

RUN chmod +x ./bot/bot

ENV BOT_TOKEN=${BOT_TOKEN}
ENTRYPOINT ["/bin/sh", "-c", "cd bot && /app/sushi-roulette/bot/bot"]
