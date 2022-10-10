# Telegram bot accounting for financial spending

Example config file ```configs/config.yml```: 

```yaml
TelegramBotToken: "<token>"

Postgres:
  username: "postgres"
  host: "localhost"
  port: "5432"
  dbname: "postgres"
  sslmode: "disable"
```
Example environment file ```.env```:

```shell
POSTGRES_PASSWORD=mysecretpassword
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