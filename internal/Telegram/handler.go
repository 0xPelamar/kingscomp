package Telegram

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

}

func (t *Telegram) start(c telebot.Context) error {
	return c.Reply("start called!!!!!!")
}
