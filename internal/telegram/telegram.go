package telegram

import (
	"errors"
	"fmt"
	"time"

	"github.com/0xpelamar/kingscomp/internal/matchmaking"
	"github.com/0xpelamar/kingscomp/internal/service"
	"github.com/0xpelamar/kingscomp/internal/telegram/teleprompt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v4"
)

type Telegram struct {
	App *service.App
	bot *telebot.Bot

	TelePrompt  *teleprompt.TelePrompt
	matchMaking matchmaking.MatchMaking
}

func NewTelegram(app *service.App, mm matchmaking.MatchMaking, apiKey string) (*Telegram, error) {
	tel := &Telegram{
		App:         app,
		TelePrompt:  teleprompt.NewTelePrompt(),
		matchMaking: mm,
	}
	pref := telebot.Settings{
		Token:   apiKey,
		Poller:  &telebot.LongPoller{Timeout: 60 * time.Second},
		OnError: tel.onError,
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		logrus.WithError(err).Fatal("failed to connect to telegram servers")
		return nil, err
	}
	tel.bot = bot

	tel.setupHandlers()
	return tel, nil
}

func (t *Telegram) onError(err error, c telebot.Context) {
	if errors.Is(err, ErrInputTimeout) || errors.Is(err, teleprompt.ErrIsCanceled) {
		return
	}
	errorID := uuid.New().String()
	logrus.WithError(err).WithField("tracing_id", errorID).Error("telegram got an error")
	_ = c.Reply(fmt.Sprintf("❌ There is an error while processing data. \ncode: %s", errorID))

}

func (t *Telegram) Start() {
	t.bot.Start()
}
