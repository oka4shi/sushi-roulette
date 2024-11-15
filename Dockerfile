FROM python:3.13.0-alpine3.20 as scraper

WORKDIR /app/sushi-roulette/scraper
COPY ./scraper/ .

RUN python3 hama-sushi.py

FROM golang:1.17.1-alpine3.13 as builder

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