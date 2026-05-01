package teleprompt

import (
	"sync"
	"time"

	"gopkg.in/telebot.v4"
)

type Prompt struct {
	TeleCtx telebot.Context
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
		close(preChannel.(chan Prompt))
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

func (t *TelePrompt) AsMessage(userID int64, timeout time.Duration) (*telebot.Message, bool) {
	ch := t.RegisterAccount(userID)
	select {
	case val := <-ch:
		return val.TeleCtx.Message(), false
	case <-time.After(timeout):
		return nil, true
	}
}
