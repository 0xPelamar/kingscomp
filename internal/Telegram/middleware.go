package Telegram

import (
	"context"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v4"
)

func (t *Telegram) registerMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		logrus.Infoln("Middleware called")
		acc := entity.Account{
			ID:          c.Sender().ID,
			FirstName:   c.Sender().FirstName,
			LastName:    c.Sender().LastName,
			Username:    c.Sender().Username,
			DisplayName: c.Sender().FirstName + " " + c.Sender().LastName,
		}
		acc, isCreated, err := t.App.Account.CreateOrUpdate(context.Background(), acc)
		c.Set("is_created", isCreated)
		c.Set("account", acc)
		if err != nil {
			return err
		}
		return next(c)
	}
}
