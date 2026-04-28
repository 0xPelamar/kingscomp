package Telegram

import "gopkg.in/telebot.v4"

func (t *Telegram) setupHandlers() {
	t.bot.Handle("/start", t.start)
}

func (t *Telegram) start(c telebot.Context) error {
	return c.Reply("start called!")
}
