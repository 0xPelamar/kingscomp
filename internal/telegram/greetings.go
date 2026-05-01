package telegram

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v4"
)

func (t *Telegram) start(c telebot.Context) error {
	isJustCreated := c.Get("is_just_created").(bool)
	if !isJustCreated {
		logrus.Infoln("It's not just created")
		return t.myInfo(c)
	}
	logrus.Infoln("It's just created")
	if err := t.editDisplayNamePrompt(c); err != nil {
		return err
	}
	return t.myInfo(c)
}

func (t *Telegram) myInfo(c telebot.Context) error {
	account := getAccount(c)
	selector := &telebot.ReplyMarkup{}
	selector.Inline(selector.Row(btnEditDisplayName))
	return c.Send(fmt.Sprintf("🏰 King «%s»\nWelcome to the Kings Combat\nWhat can I do for you?", account.DisplayName), selector)
}
