package spending

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates/nbrb"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/kafka"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/configtest"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
	"strings"
)

type ServiceTest struct {
	ctx     context.Context
	Service *Service
	DB      *sqlx.DB
	Mock    sqlmock.Sqlmock
	Cfg     *configtest.ConfigTest
}

func NewServiceTest(ctx context.Context) (st *ServiceTest, service *Service, mock sqlmock.Sqlmock, err error) {
	st = &ServiceTest{
		ctx: ctx,
	}
	st.DB, st.Mock, err = sqlmock.Newx()
	if err != nil {
		return nil, nil, nil, err
	}

	st.Cfg, _ = configtest.InitConfigTest()
	tgClient, _ := st.initTelegramBot()
	st.initMock()

	kafkaProducer, err := st.initKafkaProducer(kafka.BrokersList)
	if err != nil {
		return nil, nil, nil, err
	}
	defer kafkaProducer.Close()

	repos, _ := repository.NewRepository(st.DB)
	reposCurrencies, _ := currency.NewCurrencies(st.DB)
	ratesClient := st.initRates(repos)
	st.Service = NewService(repos.Spending, repos.Categories, reposCurrencies, tgClient, ratesClient, kafkaProducer)

	return st, st.Service, st.Mock, nil
}

func (st *ServiceTest) initMock() {
	rowsCurrency := sqlmock.NewRows([]string{"id", "abbreviation"}).
		AddRow(1, "USD").
		AddRow(2, "CNY").
		AddRow(3, "EUR").
		AddRow(4, "RUB")
	st.Mock.ExpectQuery("SELECT (.+) FROM currency").
		WillReturnRows(rowsCurrency)
}

func (st *ServiceTest) initTelegramBot() (tgClient *client.Client, err error) {
	tgBot, err := tgbotapi.NewBotAPI(st.Cfg.Token)
	if err != nil {
		return nil, errors.Wrap(err, "error init telegram bot")
	}

	tgClient, err = client.NewClient(st.ctx, tgBot)
	if err != nil {
		return nil, errors.Wrap(err, "telegram client init failed")
	}

	return
}

func (st *ServiceTest) initKafkaProducer(brokerList []string) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V3_2_3_0
	config.Producer.Return.Successes = true

	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		return nil, fmt.Errorf("starting Sarama producer: %w", err)
	}

	go func() {
		for err = range producer.Errors() {
			logger.Infos("Failed to write message:", err)
		}
	}()

	return producer, nil
}

func (st *ServiceTest) initRates(repos *repository.Repository) rates.Client {
	ratesClient := nbrb.NewRates(st.DB, repos.CurrencyClient)
	run := ratesClient.UpdateRatesSync(st.ctx)

	if run {
		go func() {
			// read channel error
			<-ratesClient.SyncChan(st.ctx)
		}()
	}

	return ratesClient
}

func (st *ServiceTest) TgBotMessageCommand(cmd string, args ...string) (update tgbotapi.Update) {
	update = tgbotapi.Update{
		Message: &tgbotapi.Message{
			MessageID: 1308,
			Chat: &tgbotapi.Chat{
				ID: st.Cfg.ChatId,
			},
		},
	}

	update.Message.Text = cmd + " " + strings.Join(args, " ")
	update.Message.Entities = &[]tgbotapi.MessageEntity{
		{
			Type:   "bot_command",
			Length: len(cmd),
		},
	}

	return
}
