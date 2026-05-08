package telegram

import (
	"time"

	"github.com/0xpelamar/kingscomp/internal/config"
	"github.com/0xpelamar/kingscomp/internal/entity"
	"gopkg.in/telebot.v4"
)

var (
	DefaultMatchMakingTimeout         = time.Second * 10
	DefaultMatchMakingLoadingInterval = time.Second * 1
	DefaultInputTimeout               = time.Minute * 5
	DefaultInputTimeoutMessage        = "We were waiting for you but you didn't send any message. Please send message when you come back."
	TxtConfirm                        = "✅ Confirm"
	TxtDecline                        = "❌ Decline"
)

var (
	selector           = &telebot.ReplyMarkup{}
	btnEditDisplayName = selector.Data("✏️ Edit Name", "btnEditName")
	btnJoinMatchMaking = selector.Data("🎮 Join new game", "btnJoinMatchMaking")
	btnCurrentMatch    = selector.Data("🕹 Current Match", "btnCurrentMatch")
	btnResignMatch     = selector.Data("🏳️ Resign Match", "btnResignMatch")
	btnStartGameWebapp = selector.Data("🎮 Starting game", "btnStartGameWebapp")
)

func getAccount(c telebot.Context) entity.Account {
	return c.Get("account").(entity.Account)
}

func NewStartWebApp(lobbyID string) telebot.Btn {
	return selector.WebApp("🎮 Opening the game", &telebot.WebApp{
		URL: config.Default.WebAppAddr + "/lobby/" + lobbyID,
	})
}
