FROM golang:1.19.2-alpine3.16 AS builder

RUN go version

COPY . /workshop-1-bot/
WORKDIR /workshop-1-bot/

RUN go mod download
RUN GOOS=linux go build -o ./.bin/bot ./cmd/bot/main.go

FROM artsafin/goose-migrations AS goose

FROM alpine:latest

WORKDIR /app

COPY --from=builder /workshop-1-bot/.bin/bot .
COPY --from=builder /workshop-1-bot/swagger swagger/
COPY --from=builder /workshop-1-bot/configs/config.yml configs/config.yml
RUN touch .env

COPY --from=builder /workshop-1-bot/migrations/*.sql migrations/
COPY --from=builder /workshop-1-bot/app.sh .

COPY --from=goose /bin/goose .

CMD /app/app.sh
