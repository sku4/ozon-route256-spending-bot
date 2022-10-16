package spending

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates/nbrb"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
	"strings"
)

type ServiceTest struct {
	ctx     context.Context
	Service *Service
	DB      *sqlx.DB
	Mock    sqlmock.Sqlmock
	Cfg     *ConfigTest
}

func NewServiceTest(ctx context.Context) (st *ServiceTest, service *Service, mock sqlmock.Sqlmock, err error) {
	st = &ServiceTest{
		ctx: ctx,
	}
	st.DB, st.Mock, err = sqlmock.Newx()
	if err != nil {
		return nil, nil, nil, err
	}

	st.Cfg, _ = initConfigTest()
	tgClient, _ := st.initTelegramBot()
	st.initMock()

	repos, _ := repository.NewRepository(st.DB)
	reposCurrencies, _ := currency.NewCurrencies(st.DB)
	ratesClient := st.initRates(repos)
	st.Service = NewService(repos.Spending, repos.Categories, reposCurrencies, tgClient, ratesClient)

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

type ConfigTest struct {
	TelegramBot `mapstructure:"Test"`
}

type TelegramBot struct {
	Token  string `mapstructure:"TelegramBotToken"`
	ChatId int64  `mapstructure:"TelegramBotChatId"`
}

func initConfigTest() (*ConfigTest, error) {
	mainViper := viper.New()
	mainViper.AddConfigPath("../../../configs")
	if err := mainViper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg ConfigTest

	if err := mainViper.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "config viper unmarshal")
	}

	return &cfg, nil
}
