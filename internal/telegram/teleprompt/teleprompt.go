package teleprompt

import (
	"errors"
	"sync"
	"time"

	"gopkg.in/telebot.v4"
)

var (
	ErrIsCanceled = errors.New("teleprompt is canceled by the user")
	ErrTimeout    = errors.New("teleprompt is timeout")
)

type Prompt struct {
	TeleCtx    telebot.Context
	IsCanceled bool
}
type TelePrompt struct {
	accountsPrompts sync.Map
}

func NewTelePrompt() *TelePrompt {
	return &TelePrompt{}
}

func (t *TelePrompt) RegisterAccount(userID int64) <-chan Prompt {
	ch := make(chan Prompt, 1)

	if preChannel, loaded := t.accountsPrompts.LoadAndDelete(userID); loaded {
		preChannel.(chan Prompt) <- Prompt{IsCanceled: true}
	}

	t.accountsPrompts.Store(userID, ch)
	return ch
}

func (t *TelePrompt) Dispatch(userID int64, c telebot.Context) bool {
	ch, loaded := t.accountsPrompts.LoadAndDelete(userID)
	if !loaded {
		return false
	}

	select {
	case ch.(chan Prompt) <- Prompt{TeleCtx: c}:
	default:
		return false
	}
	return true
}

func (t *TelePrompt) AsMessage(userID int64, timeout time.Duration) (*telebot.Message, error) {
	ch := t.RegisterAccount(userID)
	select {
	case val := <-ch:
		if val.IsCanceled {
			return nil, ErrIsCanceled
		}
		return val.TeleCtx.Message(), nil
	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}
