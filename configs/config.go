package configs

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	TelegramBotToken string `mapstructure:"TelegramBotToken"`
	ServiceName      string `mapstructure:"ServiceName"`
	RestPort         int    `mapstructure:"restPort"`
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

	return &cfg, nil
}
