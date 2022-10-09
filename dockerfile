FROM golang:1.19.2-alpine3.16 AS builder

RUN go version

COPY . /gitlab.ozon.dev/skubach/workshop-1-bot/
WORKDIR /gitlab.ozon.dev/skubach/workshop-1-bot/

RUN go mod download
RUN GOOS=linux go build -o ./.bin/app ./cmd/bot/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=0 /gitlab.ozon.dev/skubach/workshop-1-bot/.bin/app .
COPY --from=0 /gitlab.ozon.dev/skubach/workshop-1-bot/configs/config.yml configs/config.yml

CMD ["./app"]