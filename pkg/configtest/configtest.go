package configtest

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type ConfigTest struct {
	Test `mapstructure:"Test"`
}

type Test struct {
	Telegram `mapstructure:"Telegram"`
}

type Telegram struct {
	Token  string `mapstructure:"BotToken"`
	ChatId int64  `mapstructure:"ChatId"`
	UserId int    `mapstructure:"UserId"`
}

func InitConfigTest() (*ConfigTest, error) {
	mainViper := viper.New()
	mainViper.AddConfigPath("../../../configs")
	mainViper.AddConfigPath("../../configs")
	mainViper.AddConfigPath("../configs")
	mainViper.AddConfigPath("./configs")
	if err := mainViper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg ConfigTest

	if err := mainViper.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "config viper unmarshal")
	}

	return &cfg, nil
}
