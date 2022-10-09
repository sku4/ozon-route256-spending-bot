package configs

import (
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	TelegramBotToken string `mapstructure:"TelegramBotToken"`
	Username         string `mapstructure:"Postgres.username"`
	Host             string `mapstructure:"Postgres.host"`
	Port             string `mapstructure:"Postgres.port"`
	DBName           string `mapstructure:"Postgres.dbname"`
	SslMode          string `mapstructure:"Postgres.sslmode"`
	Password         string
}

func Init() (*Config, error) {
	mainViper := viper.New()
	mainViper.AddConfigPath("configs")
	if err := mainViper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config

	if err := mainViper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.Password = os.Getenv("POSTGRES_PASSWORD")

	return &cfg, nil
}
