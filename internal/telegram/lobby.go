package telegram

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/internal/matchmaking"
	"gopkg.in/telebot.v4"
)

func (t *Telegram) joinMatchMaking(c telebot.Context) error {
	msg, err := t.Input(c, InputConfig{
		Prompt:         "⏰ Each game takes about 5 minutes and if you leave the game you lose \nDo you start the game??",
		PromptKeyboard: [][]string{{TxtDecline, TxtConfirm}},
		Validator:      choiceValidator(TxtDecline, TxtConfirm),
	})
	if err != nil {
		return err
	}
	if msg.Text != TxtConfirm {
		return t.myInfo(c)
	}
	ch := make(chan struct{}, 1)
	var lobby entity.Lobby
	go func() {
		lobby, _, err = t.matchMaking.Join(context.Background(), c.Sender().ID, time.Minute*2)
		ch <- struct{}{}
	}()

	ticker := time.NewTicker(DefualtMatchMakingLoadingInterval)
	loadingMessage, err := c.Bot().Send(c.Sender(), "Searching.. please wait...")
	if err != nil {
		return err
	}
	defer func() {
		c.Bot().Delete(loadingMessage)
	}()
	s := time.Now()
loading:
	for {
		select {
		case <-ticker.C:
			took := int(time.Since(s).Seconds())
			c.Bot().Edit(loadingMessage, fmt.Sprintf("Searching for player... %d seconds took of %d", took, DefaultMatchMakingTimeout))
			continue

		case <-ch:
			break loading
		}
	}
	if err != nil {
		if errors.Is(err, matchmaking.ErrTimeout) {
			c.Send("🕛 We looked for game for 2 minutes but didn't find anything. try again later")
			return t.myInfo(c)
		}
		return err
	}

	return c.Send(fmt.Sprintf("you have joined lobby: %s", lobby.ID))
}
