package Telegram

import (
	"time"

	"github.com/0xpelamar/kingscomp/internal/service"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v4"
)

type Telegram struct {
	App *service.App
	bot *telebot.Bot
}

func NewTelegram(app *service.App, apiKey string) (*Telegram, error) {
	pref := telebot.Settings{
		Token:  apiKey,
		Poller: &telebot.LongPoller{Timeout: 60 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		logrus.WithError(err).Fatal("failed to connect to telegram servers")
		return nil, err
	}
	tel := &Telegram{
		App: app,
		bot: bot,
	}
	tel.setupHandlers()
	return tel, nil
}

func (t *Telegram) Start() {
	t.bot.Start()
}
