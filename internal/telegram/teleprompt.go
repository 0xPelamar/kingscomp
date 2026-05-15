package telegram

import (
	"errors"

	"github.com/0xpelamar/kingscomp/internal/telegram/teleprompt"
	"github.com/samber/lo"
	"gopkg.in/telebot.v4"
)

type confirm struct {
	ConfirmText func(c *telebot.Message) string
}
type InputConfig struct {
	Prompt         any
	OnTimeout      any
	PromptKeyboard [][]string
	Confirm        confirm
	Validator      Validator
}

var (
	ErrInputTimeout = errors.New("input timeout")
)

func (t *Telegram) Input(c telebot.Context, config InputConfig) (*telebot.Message, error) {
getInput:
	// This part sends a prompt to the user and asks for data
	if config.Prompt != nil {
		if config.PromptKeyboard != nil {
			c.Send(config.Prompt, generateKeyboard(config.PromptKeyboard))
		} else {
			c.Send(config.Prompt)

		}
	}
	// waits for the client until the response is fetched
	response, err := t.TelePrompt.AsMessage(c.Sender().ID, DefaultInputTimeout)
	if err != nil {
		if errors.Is(err, teleprompt.ErrTimeout) {
			if config.OnTimeout != nil {
				c.Send(config.OnTimeout)
			} else {
				c.Send(DefaultInputTimeoutMessage)
			}
			return nil, ErrInputTimeout
		}
		return nil, err

	}

	// validate
	if config.Validator.Validator != nil && !config.Validator.Validator(response) {
		c.Send(config.Validator.OnInvalid(response))
		goto getInput
	}

	// client has to confirm
	if config.Confirm.ConfirmText != nil {
		confirmText := config.Confirm.ConfirmText(response)
		confirmMessage, err := t.Input(c, InputConfig{
			Prompt:         confirmText,
			PromptKeyboard: [][]string{{TxtDecline, TxtConfirm}},
			Validator:      choiceValidator(TxtConfirm, TxtDecline),
		})
		if err != nil {
			return nil, err
		}

		// on confirm we do nothing
		if confirmMessage.Text != TxtConfirm {
			goto getInput
		}
	}
	return response, nil
}

func generateKeyboard(rows [][]string) *telebot.ReplyMarkup {
	mu := &telebot.ReplyMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		RemoveKeyboard:  true,
		ForceReply:      true,
	}

	mu.Reply(lo.Map(rows, func(row []string, _ int) telebot.Row {
		return mu.Row(lo.Map(row, func(btn string, _ int) telebot.Btn {
			return telebot.Btn{Text: btn}
		})...)
	})...)
	return mu
}
