package telegram

import (
	"context"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"gopkg.in/telebot.v4"
)

func (t *Telegram) registerMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		acc := entity.Account{
			ID:          c.Sender().ID,
			FirstName:   c.Sender().FirstName,
			LastName:    c.Sender().LastName,
			Username:    c.Sender().Username,
			DisplayName: c.Sender().FirstName + " " + c.Sender().LastName,
		}
		account, isJustCreated, err := t.App.Account.CreateOrUpdate(context.Background(), acc)
		if err != nil {
			return err
		}
		c.Set("is_just_created", isJustCreated)
		c.Set("account", account)

		return next(c)
	}
}
