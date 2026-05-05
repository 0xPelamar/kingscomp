package telegram

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v4"
)

func (t *Telegram) setupHandlers() {
	logrus.Infoln("setupHandlers called")

	// middlewares
	t.bot.Use(t.registerMiddleware)

	// handlers
	t.bot.Handle("/start", t.start)
	t.bot.Handle(telebot.OnText, t.textHandler)
	t.bot.Handle(&btnEditDisplayName, t.editDisplayName)
	t.bot.Handle(&btnJoinMatchMaking, t.joinMatchMaking)

}

func (t *Telegram) textHandler(c telebot.Context) error {
	if t.TelePrompt.Dispatch(c.Sender().ID, c) {
		return nil
	}
	// TODO: per state
	return c.Reply("I did not understand (dispatch failed)")
}
