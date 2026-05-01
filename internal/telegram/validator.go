package telegram

import (
	"slices"

	"gopkg.in/telebot.v4"
)

type Validator struct {
	Validator func(msg *telebot.Message) bool
	OnInvalid func(msg *telebot.Message) string
}

func choiceValidator(choices ...string) Validator {
	return Validator{
		Validator: func(msg *telebot.Message) bool {
			return slices.Contains(choices, msg.Text)
		},
		OnInvalid: func(msg *telebot.Message) string {
			return "Choose one of the keyboard buttons"
		},
	}
}
