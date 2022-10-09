package configs

import (
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	TelegramBotToken string `mapstructure:"TelegramBotToken"`
	Postgres         `mapstructure:"Postgres"`
}

type Postgres struct {
	Username string `mapstructure:"username"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	DBName   string `mapstructure:"dbname"`
	SslMode  string `mapstructure:"sslmode"`
	Password string
}

func Init() (*Config, error) {
	mainViper := viper.New()
	mainViper.AddConfigPath("configs")
	if err := mainViper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config

	if err := mainViper.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "config viper unmarshal")
	}

	if err := godotenv.Load(); err != nil {
		return nil, errors.Wrap(err, "error loading env variables")
	}

	cfg.Postgres.Password = os.Getenv("POSTGRES_PASSWORD")

	return &cfg, nil
}
