package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"github.com/0xpelamar/kingscomp/internal/matchmaking"
	"github.com/samber/lo"
	"gopkg.in/telebot.v4"
)

func (t *Telegram) joinMatchMaking(c telebot.Context) error {
	myAccount := getAccount(c)
	if myAccount.CurrentLobby != "" { // TODO: Show the current game status
		return c.Reply("You already have a lobby with this account")
	}
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
		lobby, _, err = t.matchMaking.Join(context.Background(), c.Sender().ID, 10*time.Second)
		ch <- struct{}{}
	}()

	ticker := time.NewTicker(DefaultMatchMakingLoadingInterval)
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
			c.Bot().Edit(loadingMessage, fmt.Sprintf("Searching for player... %d seconds took of %d", took, int(DefaultMatchMakingTimeout.Seconds())))
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
	myAccount.CurrentLobby = lobby.ID
	c.Set("account", myAccount)
	return t.currentLobby(c)

}

func (t *Telegram) currentLobby(c telebot.Context) error {
	myAccount := getAccount(c)
	lobby, accounts, err := t.App.LobbyPlayers(context.Background(), myAccount.CurrentLobby)
	if err != nil {
		return err
	}
	selector := &telebot.ReplyMarkup{}
	selector.Inline(selector.Row(btnResignMatch, NewStartWebApp(lobby.ID)))
	return c.Send(fmt.Sprintf("🏆 Your running game:\nLobby: %s\nPlayers: %s", lobby.ID, strings.Join(lo.Map(accounts, func(item entity.Account, _ int) string {
		player := fmt.Sprintf("🎴 %s %d", item.FirstName, item.ID)
		if myAccount.ID == item.ID {
			player = player + " (You)"
		}
		return player
	}), "\n")), selector)
}
