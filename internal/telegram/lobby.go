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
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v4"
)

func (t *Telegram) joinMatchMaking(c telebot.Context) error {
	myAccount := getAccount(c)
	if myAccount.CurrentLobby != "" { // TODO: Show the current lobby status
		return c.Reply("You already have a lobby with this account")
	}
	msg, err := t.Input(c, InputConfig{
		Prompt:         "⏰ Each lobby takes about 5 minutes and if you leave the lobby you lose \nDo you start the lobby??",
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
	var isHost bool
	go func() {
		lobby, isHost, err = t.matchMaking.Join(context.Background(), c.Sender().ID, 10*time.Second)
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
			c.Send("🕛 We looked for lobby for 2 minutes but didn't find anything. try again later")
			return t.myInfo(c)
		}
		return err
	}

	// Setup reminder
	if isHost {
		go func() {
			<-time.After(DefaultReminderToReadyAfter)
			lobby, err := t.App.Lobby.Get(context.Background(), lobby.EntityID())
			if err != nil {
				logrus.WithError(err).Errorln("could not get lobby to send reminder")
			}

			for _, participant := range lobby.Participants {
				if !lobby.UserState[participant].IsReady {
					c.Bot().Send(&telebot.User{ID: participant},
						"⚠️ A new game was created for but you still have not joined. please join the game",
						NewLobbyInlineKeyboard(lobby.ID),
					)
				}
			}
			<-time.After(DefaultReadyDeadline - DefaultReminderToReadyAfter)
			for _, participant := range lobby.Participants {
				if !lobby.UserState[participant].IsReady {
					t.App.Lobby.UpdateUserState(context.Background(), lobby.ID, participant, "is_resigned", true)
					if err := t.App.Account.SetField(context.Background(),
						entity.NewID("account", participant),
						"current_lobby", ""); err != nil {
						logrus.WithError(err).Errorln("could not resign the user and change current lobby")
					}
					c.Bot().Send(&telebot.User{ID: participant},
						"⚠️ We have changed your state to resigned because you have not joined the game",
						NewLobbyInlineKeyboard(lobby.ID),
					)
				}
			}
		}()
	}

	myAccount.CurrentLobby = lobby.ID
	c.Set("account", myAccount)

	return t.currentLobby(c)

}
func NewLobbyInlineKeyboard(lobbyID string) *telebot.ReplyMarkup {
	selector := &telebot.ReplyMarkup{}
	selector.Inline(selector.Row(btnResignMatch, NewStartWebApp(lobbyID)))
	return selector
}
func (t *Telegram) currentLobby(c telebot.Context) error {
	myAccount := getAccount(c)
	lobby, accounts, err := t.App.LobbyPlayers(context.Background(), myAccount.CurrentLobby)
	if err != nil {
		return err
	}
	selector := &telebot.ReplyMarkup{}
	selector.Inline(selector.Row(btnResignMatch, NewStartWebApp(lobby.ID)))
	return c.Send(fmt.Sprintf("🏆 Your running lobby:\nLobby: %s\nPlayers: %s", lobby.ID, strings.Join(lo.Map(accounts, func(item entity.Account, _ int) string {
		player := fmt.Sprintf("🎴 %s %d", item.FirstName, item.ID)
		if myAccount.ID == item.ID {
			player = player + " (You)"
		}
		return player
	}), "\n")), selector)
}
func (t *Telegram) resignLobby(c telebot.Context) error {
	myAccount := getAccount(c)
	myAccount.CurrentLobby = ""
	if err := t.App.Account.Save(context.Background(), myAccount); err != nil {
		return err
	}
	c.Send("✅ You resigned successfully")
	c.Set("account", myAccount)
	return t.myInfo(c)
}
