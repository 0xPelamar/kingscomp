package telegram

import (
	"errors"

	"github.com/0xpelamar/kingscomp/internal/telegram/teleprompt"
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
	}
	telRows := make([]telebot.Row, 0, len(rows))
	for _, row := range rows {
		telBtns := make([]telebot.Btn, 0, len(row))
		for _, btn := range row {
			telBtn := telebot.Btn{
				Text: btn,
			}
			telBtns = append(telBtns, telBtn)
		}
		telRows = append(telRows, telBtns)
	}
	mu.Reply(telRows...)
	return mu
}
