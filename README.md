# Telegram bot accounting for financial spending

System design on [miro.com](https://miro.com/app/board/uXjVPLT339I=/)

Example config file ```configs/config.yml```: 

```yaml
TelegramBotToken: "<token>"

Test:
  Telegram:
    BotToken: "<test_token>"
    ChatId: 123456789
    UserId: 123456789
```
Example environment file ```.env```:

```shell
POSTGRES_HOST=localhost
POSTGRES_PORT=5433
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB_NAME=postgres
POSTGRES_SSL=disable
```
## Available commands:
- /categories
- `/categoryadd Food` - where Food is category name
- `/spendingadd 100` - where 100 is price
- /report7 - report by current week
- /report31 - report by current month
- /report365 - report by current year
- /currency - change currency
- `/limit 100` - limit category by sum spending on month
### Run app:

```
make docker-run
```

Before first exec:

```
make goose-up
```
## Integration tests:
```
docker compose up
go test ./... -tags integration
```