//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	main "gitlab.ozon.dev/skubach/workshop-1-bot/cmd/bot"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/handler"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service/spending"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/configtest"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/decimal"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type SpendingSuite struct {
	suite.Suite
	ctx context.Context
	cfg *configtest.ConfigTest
	db  *sqlx.DB
	hr  *handler.Handler
}

func (s *SpendingSuite) SetupSuite() {
	s.ctx = context.Background()
	s.cfg, _ = configtest.InitConfigTest()
	tgClient := s.initTelegramBot()

	err := godotenv.Load("../.env")
	s.Require().NoError(err)

	s.db, err = postgres.NewPostgresDB(postgres.Config{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Username: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB_NAME"),
		SslMode:  os.Getenv("POSTGRES_SSL"),
	})
	s.Require().NoError(err)

	repos, err := repository.NewRepository(s.db)
	s.Require().NoError(err)
	ratesClient := main.InitRates(s.ctx, s.db, repos)
	services := service.NewService(repos, tgClient, ratesClient)
	s.hr = handler.NewHandler(s.ctx, services)
}

func (s *SpendingSuite) TestCategoryAdd() {
	categoryTitle := "Test category"
	tgBotUpdateCommand := s.tgBotMessageCommand("/categoryadd", categoryTitle)
	err := s.hr.IncomingMessage(tgBotUpdateCommand)
	s.Require().NoError(err)

	var c model.Category
	query := fmt.Sprintf(`SELECT id, title FROM category WHERE title=$1`)
	err = s.db.GetContext(s.ctx, &c, query, categoryTitle)
	s.Require().NoError(err)

	s.Require().Equal(c.Title, categoryTitle)
}

func (s *SpendingSuite) TestEventAdd() {
	price := float64(100)
	priceDecimal := decimal.ToDecimal(price)
	categoryId := 1
	messageId := 123456
	tgBotUpdateCommand := s.tgBotMessageCommand("/spendingadd", strconv.FormatFloat(price, 'f', 0, 64))
	tgBotUpdateCommand.Message.MessageID = messageId
	err := s.hr.IncomingMessage(tgBotUpdateCommand)
	s.Require().NoError(err)

	event := spending.NewEvent(price)
	event.CategoryId = categoryId
	event.D = 12
	event.M = 5
	event.Y = 1990
	event.Today = true
	event.SelectedToday = true
	eventSer := spending.AddPrefix + spending.EventSerialize(event)
	tgBotUpdateQuery := s.tgBotCallbackQuery(eventSer)
	tgBotUpdateQuery.CallbackQuery.Message.MessageID = messageId
	_ = s.hr.IncomingMessage(tgBotUpdateQuery)

	now := time.Now().UTC()
	f1 := time.Date(1990, 5, 12, 0, 0, 0, 0, now.Location())
	f2 := f1.AddDate(0, 0, 1)
	var events []model.EventDB
	query := fmt.Sprintf(`SELECT id, category_id, event_at, price FROM %s WHERE event_at BETWEEN '%s' AND '%s'`,
		"event", f1.Format("2006-01-02"), f2.Format("2006-01-02"))
	err = s.db.SelectContext(s.ctx, &events, query)
	s.Require().NoError(err)

	eventFound := false
	for _, e := range events {
		if e.CategoryId == categoryId && decimal.Decimal(e.Price) == priceDecimal {
			eventFound = true
			break
		}
	}

	if !eventFound {
		s.Require().FailNow("event not found")
	}
}

func (s *SpendingSuite) initTelegramBot() (tgClient *client.Client) {
	tgBot, err := tgbotapi.NewBotAPI(s.cfg.Token)
	s.Require().NoError(err)

	tgClient, err = client.NewClient(s.ctx, tgBot)
	s.Require().NoError(err)

	return
}

func (s *SpendingSuite) tgBotMessageCommand(cmd string, args ...string) (update tgbotapi.Update) {
	var seed = time.Now().UnixNano()
	rand.Seed(seed)
	update = tgbotapi.Update{
		Message: &tgbotapi.Message{
			MessageID: int(seed),
			From: &tgbotapi.User{
				ID: s.cfg.UserId,
			},
			Chat: &tgbotapi.Chat{
				ID: s.cfg.ChatId,
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

func (s *SpendingSuite) tgBotCallbackQuery(data string) (update tgbotapi.Update) {
	var seed = time.Now().UnixNano()
	rand.Seed(seed)
	update = tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{
				MessageID: int(seed),
				From: &tgbotapi.User{
					ID: s.cfg.UserId,
				},
				Chat: &tgbotapi.Chat{
					ID: s.cfg.ChatId,
				},
			},
			From: &tgbotapi.User{
				ID: s.cfg.UserId,
			},
			Data: data,
		},
	}

	return
}

func TestSpendingSuite(t *testing.T) {
	suite.Run(t, new(SpendingSuite))
}
