package telegram

import (
	"time"

	"github.com/0xpelamar/kingscomp/internal/entity"
	"gopkg.in/telebot.v4"
)

var (
	DefaultMatchMakingTimeout         = time.Second * 10
	DefualtMatchMakingLoadingInterval = DefaultMatchMakingTimeout / 10
	DefaultInputTimeout               = time.Minute * 5
	DefaultInputTimeoutMessage        = "We were waiting for you but you didn't send any message. Please send message when you come back."
	TxtConfirm                        = "✅ Confirm"
	TxtDecline                        = "❌ Decline"
)

var (
	selector           = &telebot.ReplyMarkup{}
	btnEditDisplayName = selector.Data("✏️ Edit Name", "btnEditName")
	btnJoinMatchMaking = selector.Data("🎮 Join new game", "btnJoinMatchMaking")
	btnEditProvince    = selector.Data("✏️ Edit Province", "editProvince")
	btnEditAge         = selector.Data("✏️ Edit Age", "editAge")
	btnEditGender      = selector.Data("✏️ Edit Gender", "editGender")
)

func getAccount(c telebot.Context) entity.Account {
	return c.Get("account").(entity.Account)
}
