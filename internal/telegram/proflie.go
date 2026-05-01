package telegram

import (
	"context"
	"fmt"

	"gopkg.in/telebot.v4"
)

func (t *Telegram) editDisplayName(c telebot.Context) error {
	c.Delete()
	t.editDisplayNamePrompt(c)
	return t.myInfo(c)

}

func (t *Telegram) editDisplayNamePrompt(c telebot.Context) error {
	account := getAccount(c)
	msg, err := t.Input(c, InputConfig{
		Prompt: " 👋 Welcome to Kings Combat\nPlease enter your display name. this can be changed later.",
		Confirm: confirm{
			ConfirmText: func(msg *telebot.Message) string {
				return fmt.Sprintf("ℹ️ We call you «%s»\nDo you confirm?", msg.Text)
			},
		},
		Validator: Validator{
			Validator: func(msg *telebot.Message) bool {
				l := len([]rune(msg.Text))
				return (2 < l) && (l < 21)
			},
			OnInvalid: func(msg *telebot.Message) string {
				return "❌ Your name must be more than 3 and less than 21 characters"
			},
		},
	})
	if err != nil {
		return err
	}

	displayName := msg.Text
	// TODO validation
	account.DisplayName = displayName
	if err := t.App.Account.Update(context.Background(), account); err != nil {
		return err
	}
	c.Set("account", account)
	_ = c.Reply(fmt.Sprintf("✅ So we will call you «%s»", displayName))
	return nil
}
